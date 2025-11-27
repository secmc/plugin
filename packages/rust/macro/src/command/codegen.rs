use heck::ToSnakeCase;
use proc_macro2::TokenStream;
use quote::{format_ident, quote};
use syn::{Attribute, DeriveInput, Ident, LitStr};

use crate::command::{
    model::{collect_command_shape, CommandShape, ParamMeta, VariantMeta},
    parse::CommandInfoParser,
};

pub fn generate_command_impls(ast: &DeriveInput, attr: &Attribute) -> TokenStream {
    let command_info = match attr.parse_args::<CommandInfoParser>() {
        Ok(info) => info,
        Err(e) => return e.to_compile_error(),
    };

    let cmd_ident = &ast.ident;
    let cmd_name_lit = &command_info.name;
    let cmd_desc_lit = &command_info.description;
    let aliases_lits = &command_info.aliases;

    let shape = match collect_command_shape(ast) {
        Ok(s) => s,
        Err(e) => return e.to_compile_error(),
    };

    let spec_impl =
        generate_spec_impl(cmd_ident, cmd_name_lit, cmd_desc_lit, aliases_lits, &shape);
    let try_from_impl = generate_try_from_impl(cmd_ident, cmd_name_lit, aliases_lits, &shape);
    let trait_impl = generate_handler_trait(cmd_ident, &shape);
    let exec_impl = generate_execute_impl(cmd_ident, &shape);

    quote! {
        #spec_impl
        #try_from_impl
        #trait_impl
        #exec_impl
    }
}

fn generate_spec_impl(
    cmd_ident: &Ident,
    cmd_name_lit: &LitStr,
    cmd_desc_lit: &LitStr,
    aliases_lits: &[LitStr],
    shape: &CommandShape,
) -> TokenStream {
    let params_tokens = match shape {
        CommandShape::Struct { params } => {
            let specs = params.iter().map(param_to_spec);
            quote! { vec![ #( #specs ),* ] }
        }
        CommandShape::Enum { variants } => {
            // First param: subcommand enum
            let variant_names_iter = variants.iter().map(|v| &v.canonical);
            let subcommand_names: Vec<&LitStr> = variants
                .iter()
                .flat_map(|v| &v.aliases)
                .chain(variant_names_iter)
                .collect();
            let subcommand_spec = quote! {
                dragonfly_plugin::types::ParamSpec {
                    name: "subcommand".to_string(),
                    r#type: dragonfly_plugin::types::ParamType::ParamEnum as i32,
                    optional: false,
                    suffix: String::new(),
                    enum_values: vec![ #( #subcommand_names.to_string() ),* ],
                }
            };

            // Merge all variant params, marking as optional if not present in all variants
            let merged = merge_variant_params(variants);
            let merged_specs = merged.iter().map(param_to_spec);

            quote! { vec![ #subcommand_spec, #( #merged_specs ),* ] }
        }
    };

    quote! {
        impl #cmd_ident {
            pub fn spec() -> dragonfly_plugin::types::CommandSpec {
                dragonfly_plugin::types::CommandSpec {
                    name: #cmd_name_lit.to_string(),
                    description: #cmd_desc_lit.to_string(),
                    aliases: vec![ #( #aliases_lits.to_string() ),* ],
                    params: #params_tokens,
                }
            }
        }
    }
}

fn param_to_spec(p: &ParamMeta) -> TokenStream {
    let name_str = p.field_ident.to_string();
    let param_type_expr = &p.param_type_expr;
    let is_optional = p.is_optional;

    quote! {
        dragonfly_plugin::types::ParamSpec {
            name: #name_str.to_string(),
            r#type: #param_type_expr as i32,
            optional: #is_optional,
            suffix: String::new(),
            enum_values: Vec::new(),
        }
    }
}

/// Merge params from all variants: a param is optional if not present in every variant.
fn merge_variant_params(variants: &[VariantMeta]) -> Vec<ParamMeta> {
    use std::collections::HashMap;

    // Collect all unique param names with their metadata
    let mut seen: HashMap<String, (ParamMeta, usize)> = HashMap::new(); // name -> (meta, count)

    for variant in variants {
        for param in &variant.params {
            let name = param.field_ident.to_string();
            seen.entry(name)
                .and_modify(|(_, count)| *count += 1)
                .or_insert_with(|| (param.clone(), 1));
        }
    }

    let variant_count = variants.len();

    seen.into_iter()
        .map(|(_, (mut meta, count))| {
            // If not present in all variants, mark as optional
            if count < variant_count {
                meta.is_optional = true;
            }
            meta
        })
        .collect()
}

fn generate_try_from_impl(
    cmd_ident: &Ident,
    cmd_name_lit: &LitStr,
    cmd_aliases: &[LitStr],
    shape: &CommandShape,
) -> TokenStream {
    let body = match shape {
        CommandShape::Struct { params } => {
            let field_inits = params.iter().map(|p| struct_field_init(p, 0));
            quote! {
                Ok(Self {
                    #( #field_inits, )*
                })
            }
        }
        CommandShape::Enum { variants } => {
            let match_arms = variants.iter().map(|v| {
                let variant_ident = &v.ident;
                let params = &v.params;

                // Build patterns: "canonical" | "alias1" | "alias2"
                let mut name_lits = Vec::new();
                name_lits.push(&v.canonical);
                for alias in &v.aliases {
                    name_lits.push(alias);
                }

                let subcommand_patterns = quote! { #( #name_lits )|* };

                if params.is_empty() {
                    quote! {
                        #subcommand_patterns => Ok(Self::#variant_ident),
                    }
                } else {
                    let field_inits = params.iter().map(enum_field_init);
                    quote! {
                        #subcommand_patterns => Ok(Self::#variant_ident {
                            #( #field_inits, )*
                        }),
                    }
                }
            });

            quote! {
                let subcommand = event.args.first()
                    .ok_or(dragonfly_plugin::command::CommandParseError::Missing("subcommand"))?
                    .as_str();

                match subcommand {
                    #( #match_arms )*
                    _ => Err(dragonfly_plugin::command::CommandParseError::UnknownSubcommand),
                }
            }
        }
    };

    let mut conditions = Vec::with_capacity(1 + cmd_aliases.len());
    conditions.push(quote! { event.command != #cmd_name_lit });

    for alias in cmd_aliases {
        conditions.push(quote! { && event.command != #alias });
    }

    quote! {
        impl ::core::convert::TryFrom<&dragonfly_plugin::types::CommandEvent> for #cmd_ident {
            type Error = dragonfly_plugin::command::CommandParseError;

            fn try_from(event: &dragonfly_plugin::types::CommandEvent) -> Result<Self, Self::Error> {
                if #(#conditions)* {
                    return Err(dragonfly_plugin::command::CommandParseError::NoMatch);
                }

                #body
            }
        }
    }
}

/// Generate field init for struct commands (args start at index 0).
fn struct_field_init(p: &ParamMeta, offset: usize) -> TokenStream {
    let ident = &p.field_ident;
    let idx = p.index + offset;
    let name_str = ident.to_string();
    let ty = &p.field_ty;

    if p.is_optional {
        quote! {
            #ident: dragonfly_plugin::command::parse_optional_arg::<#ty>(&event.args, #idx, #name_str)?
        }
    } else {
        quote! {
            #ident: dragonfly_plugin::command::parse_required_arg::<#ty>(&event.args, #idx, #name_str)?
        }
    }
}

/// Generate field init for enum variant (args start at index 1, after subcommand).
fn enum_field_init(p: &ParamMeta) -> TokenStream {
    let ident = &p.field_ident;
    let idx = p.index + 1; // +1 because index 0 is the subcommand
    let name_str = ident.to_string();
    let ty = &p.field_ty;

    if p.is_optional {
        quote! {
            #ident: dragonfly_plugin::command::parse_optional_arg::<#ty>(&event.args, #idx, #name_str)?
        }
    } else {
        quote! {
            #ident: dragonfly_plugin::command::parse_required_arg::<#ty>(&event.args, #idx, #name_str)?
        }
    }
}

fn generate_handler_trait(cmd_ident: &Ident, shape: &CommandShape) -> TokenStream {
    let trait_ident = format_ident!("{}Handler", cmd_ident);

    match shape {
        CommandShape::Struct { params } => {
            // method name = struct name in snake_case, e.g. Ping -> ping
            let method_ident = format_ident!("{}", cmd_ident.to_string().to_snake_case());

            let args = params.iter().map(|p| {
                let ident = &p.field_ident;
                let ty = &p.field_ty;
                quote! { #ident: #ty }
            });

            quote! {
                #[allow(async_fn_in_trait)]
                pub trait #trait_ident: Send + Sync {
                    async fn #method_ident(
                        &self,
                        ctx: dragonfly_plugin::command::Ctx<'_>,
                        #( #args ),*
                    );
                }
            }
        }
        CommandShape::Enum { variants } => {
            let methods = variants.iter().map(|v| {
                let method_ident = format_ident!("{}", v.ident.to_string().to_snake_case());
                let args = v.params.iter().map(|p| {
                    let ident = &p.field_ident;
                    let ty = &p.field_ty;
                    quote! { #ident: #ty }
                });
                quote! {
                    async fn #method_ident(
                        &self,
                        ctx: dragonfly_plugin::command::Ctx<'_>,
                        #( #args ),*
                    );
                }
            });

            quote! {
                #[allow(async_fn_in_trait)]
                pub trait #trait_ident: Send + Sync {
                    #( #methods )*
                }
            }
        }
    }
}

fn generate_execute_impl(cmd_ident: &Ident, shape: &CommandShape) -> TokenStream {
    let trait_ident = format_ident!("{}Handler", cmd_ident);

    match shape {
        CommandShape::Struct { params } => {
            let method_ident = format_ident!("{}", cmd_ident.to_string().to_snake_case());
            let field_idents: Vec<_> = params.iter().map(|p| &p.field_ident).collect();

            quote! {
                impl #cmd_ident {
                    pub async fn __execute<H: #trait_ident>(
                        self,
                        handler: &H,
                        ctx: dragonfly_plugin::command::Ctx<'_>,
                    ) {
                        let Self { #( #field_idents ),* } = self;
                        handler.#method_ident(ctx, #( #field_idents ),*).await;
                    }
                }
            }
        }
        CommandShape::Enum { variants } => {
            let match_arms = variants.iter().map(|v| {
                let variant_ident = &v.ident;
                let method_ident = format_ident!("{}", v.ident.to_string().to_snake_case());
                let field_idents: Vec<_> = v.params.iter().map(|p| &p.field_ident).collect();

                if field_idents.is_empty() {
                    quote! {
                        Self::#variant_ident => handler.#method_ident(ctx).await,
                    }
                } else {
                    quote! {
                        Self::#variant_ident { #( #field_idents ),* } => {
                            handler.#method_ident(ctx, #( #field_idents ),*).await
                        }
                    }
                }
            });

            quote! {
                impl #cmd_ident {
                    pub async fn __execute<H: #trait_ident>(
                        self,
                        handler: &H,
                        ctx: dragonfly_plugin::command::Ctx<'_>,
                    ) {
                        match self {
                            #( #match_arms )*
                        }
                    }
                }
            }
        }
    }
}


