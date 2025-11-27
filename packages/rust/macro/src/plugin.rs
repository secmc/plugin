use quote::quote;
use syn::{
    parse::{Parse, ParseStream},
    punctuated::Punctuated,
    Attribute, Expr, Ident, Lit, LitStr, Meta, Token,
};

struct PluginInfoParser {
    pub id: LitStr,
    pub name: LitStr,
    pub version: LitStr,
    pub api: LitStr,
    pub commands: Vec<Ident>,
}

impl Parse for PluginInfoParser {
    fn parse(input: ParseStream) -> syn::Result<Self> {
        let metas = Punctuated::<Meta, Token![,]>::parse_terminated(input)?;

        let mut id = None;
        let mut name = None;
        let mut version = None;
        let mut api = None;
        let mut commands = Vec::new();

        for meta in metas {
            match meta {
                Meta::NameValue(nv) => {
                    let key_ident = nv.path.get_ident().ok_or_else(|| {
                        syn::Error::new_spanned(&nv.path, "Expected an identifier (e.g., 'id')")
                    })?;

                    let value_str = match &nv.value {
                        Expr::Lit(expr_lit) => match &expr_lit.lit {
                            // clone out the reference as we will now
                            // be owning it and putting it into our generated code
                            Lit::Str(lit_str) => lit_str.clone(),
                            _ => {
                                return Err(syn::Error::new_spanned(
                                    &nv.value,
                                    "Expected a string literal",
                                ));
                            }
                        },
                        _ => {
                            return Err(syn::Error::new_spanned(
                                &nv.value,
                                "Expected a string literal",
                            ));
                        }
                    };

                    // Store the value
                    if key_ident == "id" {
                        id = Some(value_str);
                    } else if key_ident == "name" {
                        name = Some(value_str);
                    } else if key_ident == "version" {
                        version = Some(value_str);
                    } else if key_ident == "api" {
                        api = Some(value_str);
                    } else {
                        return Err(syn::Error::new_spanned(
                            key_ident,
                            "Unknown key. Expected 'id', 'name', 'version', or 'api'",
                        ));
                    }
                }
                Meta::List(list) if list.path.is_ident("commands") => {
                    // Parse commands(Ident, Ident, ...)
                    commands = list
                        .parse_args_with(Punctuated::<Ident, Token![,]>::parse_terminated)?
                        .into_iter()
                        .collect();
                }
                _ => {
                    return Err(syn::Error::new_spanned(
                        meta,
                        "Expected `key = \"value\"` or `commands(...)` format",
                    ));
                }
            };
        }

        // Validate that all required fields were found
        // We use `input.span()` to point the error at the whole `#[plugin(...)]`
        // attribute if a field is missing.
        let id = id.ok_or_else(|| syn::Error::new(input.span(), "Missing required field 'id'"))?;
        let name =
            name.ok_or_else(|| syn::Error::new(input.span(), "Missing required field 'name'"))?;
        let version = version
            .ok_or_else(|| syn::Error::new(input.span(), "Missing required field 'version'"))?;
        let api =
            api.ok_or_else(|| syn::Error::new(input.span(), "Missing required field 'api'"))?;

        Ok(Self {
            id,
            name,
            version,
            api,
            commands,
        })
    }
}

pub(crate) fn generate_plugin_impl(
    attr: &Attribute,
    derive_name: &Ident,
) -> proc_macro2::TokenStream {
    let plugin_info = match attr.parse_args::<PluginInfoParser>() {
        Ok(info) => info,
        Err(e) => return e.to_compile_error(),
    };

    // gotta define these outside because of quote rules with the . access.
    let id_lit = &plugin_info.id;
    let name_lit = &plugin_info.name;
    let version_lit = &plugin_info.version;
    let api_lit = &plugin_info.api;

    let commands = &plugin_info.commands;

    let command_registry_impl = if commands.is_empty() {
        // No commands: empty impl uses default (returns empty vec)
        quote! {
            impl dragonfly_plugin::command::CommandRegistry for #derive_name {}
        }
    } else {
        // Has commands: generate get_commands() and dispatch_commands()
        let spec_calls = commands.iter().map(|cmd| {
            quote! { #cmd::spec() }
        });

        let dispatch_arms = commands.iter().map(|cmd| {
            quote! {
                match #cmd::try_from(event.data) {
                    Ok(cmd) => {
                        event.cancel().await;
                        cmd.__execute(self, ctx).await;
                        return true;
                    }
                    Err(dragonfly_plugin::command::CommandParseError::NoMatch) => {
                        // Try the next registered command.
                    }
                    Err(
                        err @ dragonfly_plugin::command::CommandParseError::Missing(_)
                        | err @ dragonfly_plugin::command::CommandParseError::Invalid(_)
                        | err @ dragonfly_plugin::command::CommandParseError::UnknownSubcommand,
                    ) => {
                        // Surface parse errors to the player as a friendly message.
                        let _ = ctx.reply(err.to_string()).await;
                        return true;
                    }
                }
            }
        });

        quote! {
            impl dragonfly_plugin::command::CommandRegistry for #derive_name {
                fn get_commands(&self) -> Vec<dragonfly_plugin::types::CommandSpec> {
                    vec![ #( #spec_calls ),* ]
                }

                async fn dispatch_commands(
                    &self,
                    server: &dragonfly_plugin::Server,
                    event: &mut dragonfly_plugin::event::EventContext<'_, dragonfly_plugin::types::CommandEvent>,
                ) -> bool {
                    use ::core::convert::TryFrom;
                    let ctx = dragonfly_plugin::command::Ctx::new(server, event.data.player_uuid.clone());

                    #( #dispatch_arms )*

                    false
                }
            }
        }
    };

    quote! {
        impl dragonfly_plugin::Plugin for #derive_name {
            fn get_info(&self) -> dragonfly_plugin::PluginInfo<'static> {
                dragonfly_plugin::PluginInfo::<'static> {
                    id: #id_lit,
                    name: #name_lit,
                    version: #version_lit,
                    api_version: #api_lit
                }
            }
            fn get_id(&self) -> &'static str { #id_lit }
            fn get_name(&self) -> &'static str { #name_lit }
            fn get_version(&self) -> &'static str { #version_lit }
            fn get_api_version(&self) -> &'static str { #api_lit }
        }

        #command_registry_impl
    }
}
