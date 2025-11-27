use heck::ToSnakeCase;
use proc_macro2::TokenStream;
use quote::quote;
use syn::{
    Data, DeriveInput, Fields, GenericArgument, Ident, LitStr, PathArguments, Type, TypePath,
};

use crate::command::parse::{parse_subcommand_attr, SubcommandAttr};

/// Metadata for a single command parameter (struct field or enum variant field).
#[derive(Clone)]
pub(crate) struct ParamMeta {
    /// Rust field identifier, e.g. `amount`
    pub field_ident: Ident,
    /// Full Rust type of the field, e.g. `f64` or `Option<String>`
    pub field_ty: Type,
    /// Expression for ParamType (e.g. `ParamType::ParamFloat`)
    pub param_type_expr: TokenStream,
    /// Whether this is optional in the spec
    pub is_optional: bool,
    /// Position index in args (0-based, relative to this variant/struct)
    pub index: usize,
}

/// Metadata for a single enum variant (subcommand).
pub(crate) struct VariantMeta {
    /// Variant identifier, e.g. `Pay`
    pub ident: Ident,
    /// name for matching, e.g. `"pay"`
    pub canonical: LitStr,
    /// Aliases e.g give, donate.
    pub aliases: Vec<LitStr>,
    /// Parameters (fields) for this variant
    pub params: Vec<ParamMeta>,
}

/// The shape of a command: either a struct or an enum with variants.
pub(crate) enum CommandShape {
    Struct { params: Vec<ParamMeta> },
    Enum { variants: Vec<VariantMeta> },
}

pub(crate) fn collect_command_shape(ast: &DeriveInput) -> syn::Result<CommandShape> {
    match &ast.data {
        Data::Struct(data) => {
            let params = collect_params_from_fields(&data.fields)?;
            Ok(CommandShape::Struct { params })
        }
        Data::Enum(data) => {
            let mut variants_meta = Vec::new();

            for variant in &data.variants {
                let ident = variant.ident.clone();
                let default_name =
                    LitStr::new(ident.to_string().to_snake_case().as_str(), ident.span());

                let SubcommandAttr { name, aliases } = parse_subcommand_attr(&variant.attrs)?;
                let canonical = name.unwrap_or(default_name);
                let params = collect_params_from_fields(&variant.fields)?;

                variants_meta.push(VariantMeta {
                    ident,
                    canonical,
                    aliases,
                    params,
                });
            }

            if variants_meta.is_empty() {
                return Err(syn::Error::new_spanned(
                    &ast.ident,
                    "enum commands must have at least one variant",
                ));
            }

            Ok(CommandShape::Enum {
                variants: variants_meta,
            })
        }
        Data::Union(_) => Err(syn::Error::new_spanned(
            ast,
            "unions are not supported for #[derive(Command)]",
        )),
    }
}

fn collect_params_from_fields(fields: &Fields) -> syn::Result<Vec<ParamMeta>> {
    let fields = match fields {
        Fields::Named(named) => &named.named,
        Fields::Unit => {
            // No params
            return Ok(Vec::new());
        }
        Fields::Unnamed(_) => {
            return Err(syn::Error::new_spanned(
                fields,
                "tuple structs are not supported for commands; use named fields",
            ));
        }
    };

    let mut out = Vec::new();
    for (index, field) in fields.iter().enumerate() {
        let field_ident = field
            .ident
            .clone()
            .expect("command struct fields must be named");
        let field_ty = field.ty.clone();

        let (param_type_expr, is_optional) = get_param_type(&field_ty);

        out.push(ParamMeta {
            field_ident,
            field_ty,
            param_type_expr,
            is_optional,
            index,
        });
    }

    Ok(out)
}

fn get_param_type(ty: &Type) -> (TokenStream, bool) {
    if let Some(inner) = option_inner(ty) {
        let (inner_param, _inner_opt) = get_param_type(inner);
        return (inner_param, true);
    }

    if let Type::Reference(r) = ty {
        return get_param_type(&r.elem);
    }

    if let Type::Path(TypePath { path, .. }) = ty {
        if let Some(seg) = path.segments.last() {
            let ident = seg.ident.to_string();

            // Floats
            if ident == "f32" || ident == "f64" {
                return (
                    quote! { dragonfly_plugin::types::ParamType::ParamFloat },
                    false,
                );
            }

            if matches!(
                ident.as_str(),
                "i8" | "i16"
                    | "i32"
                    | "i64"
                    | "i128"
                    | "u8"
                    | "u16"
                    | "u32"
                    | "u64"
                    | "u128"
                    | "isize"
                    | "usize"
            ) {
                return (
                    quote! { dragonfly_plugin::types::ParamType::ParamInt },
                    false,
                );
            }

            if ident == "bool" {
                return (
                    quote! { dragonfly_plugin::types::ParamType::ParamBool },
                    false,
                );
            }

            if ident == "String" {
                return (
                    quote! { dragonfly_plugin::types::ParamType::ParamString },
                    false,
                );
            }
        }
    }

    (
        quote! { dragonfly_plugin::types::ParamType::ParamString },
        false,
    )
}

fn option_inner(ty: &Type) -> Option<&Type> {
    if let Type::Path(TypePath { path, .. }) = ty {
        if let Some(seg) = path.segments.last() {
            if seg.ident == "Option" {
                if let PathArguments::AngleBracketed(args) = &seg.arguments {
                    for arg in &args.args {
                        if let GenericArgument::Type(inner) = arg {
                            return Some(inner);
                        }
                    }
                }
            }
        }
    }
    None
}
