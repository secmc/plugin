mod command;
mod plugin;

use heck::ToPascalCase;
use proc_macro::TokenStream;
use quote::{format_ident, quote};
use syn::{
    parse_macro_input, Attribute, DeriveInput, FnArg, Ident, ImplItem, ItemImpl, Pat, PatType, Type,
};

use crate::{command::generate_command_impls, plugin::generate_plugin_impl};

#[proc_macro_derive(Plugin, attributes(plugin, events))]
pub fn handler_derive(input: TokenStream) -> TokenStream {
    let ast = parse_macro_input!(input as DeriveInput);

    let derive_name = &ast.ident;

    let info_attr = match find_attribute(
        &ast,
        "plugin",
        "Missing `#[plugin(...)]` attribute with metadata.",
    ) {
        Ok(attr) => attr,
        Err(e) => return e.to_compile_error().into(),
    };

    let plugin_impl = generate_plugin_impl(info_attr, derive_name);

    quote! {
        #plugin_impl
    }
    .into()
}

fn find_attribute<'a>(
    ast: &'a syn::DeriveInput,
    name: &str,
    error: &str,
) -> Result<&'a Attribute, syn::Error> {
    ast.attrs
        .iter()
        .find(|a| a.path().is_ident(name))
        .ok_or_else(|| syn::Error::new(ast.ident.span(), error))
}

#[proc_macro_attribute]
pub fn event_handler(_attr: TokenStream, item: TokenStream) -> TokenStream {
    let item_clone = item.clone();

    // Try to parse the input tokens as an `impl` block
    let impl_block = match syn::parse::<ItemImpl>(item) {
        Ok(block) => block,
        Err(_) => {
            // Parse failed, which means the user is probably in the
            // middle of typing. Return the original, un-parsed tokens
            // to keep the LSP alive
            return item_clone;
        }
    };

    // ensure its our EventHandler.
    let is_event_handler_impl = if let Some((_, trait_path, _)) = &impl_block.trait_ {
        trait_path
            .segments
            .last()
            .is_some_and(|segment| segment.ident == "EventHandler")
    } else {
        return item_clone;
    };

    if !is_event_handler_impl {
        // This is an `impl` for some *other* trait.
        // We shouldn't touch it. Return the original tokens.
        return item_clone;
    }

    let mut event_variants = Vec::new();
    for item in &impl_block.items {
        if let ImplItem::Fn(method) = item {
            let fn_name = method.sig.ident.to_string();

            if let Some(event_name_snake) = fn_name.strip_prefix("on_") {
                let event_name_pascal = event_name_snake.to_pascal_case();

                let variant_ident = format_ident!("{}", event_name_pascal);
                event_variants.push(quote! { types::EventType::#variant_ident });
            }
        }
    }

    let self_ty = &impl_block.self_ty;

    let subscriptions_impl = quote! {
        impl dragonfly_plugin::EventSubscriptions for #self_ty {
            fn get_subscriptions(&self) -> Vec<types::EventType> {
                vec![
                    #( #event_variants ),*
                ]
            }
        }
    };

    let original_impl_tokens = quote! { #impl_block };

    let final_output = quote! {
        #original_impl_tokens
        #subscriptions_impl
    };

    final_output.into()
}

#[proc_macro_derive(Command, attributes(command))]
pub fn command_derive(input: TokenStream) -> TokenStream {
    let ast = parse_macro_input!(input as DeriveInput);

    let command_attr = match find_attribute(
        &ast,
        "command",
        "Missing `#[command(...)]` attribute with metadata.",
    ) {
        Ok(attr) => attr,
        Err(e) => return e.to_compile_error().into(),
    };

    generate_command_impls(&ast, command_attr).into()
}

#[proc_macro_attribute]
pub fn command_handlers(_attr: TokenStream, item: TokenStream) -> TokenStream {
    let item_clone = item.clone();

    let impl_block = match syn::parse::<ItemImpl>(item) {
        Ok(block) => block,
        Err(_) => return item_clone, // Keep LSP happy during typing
    };

    // Must be an inherent impl (not a trait impl)
    if impl_block.trait_.is_some() {
        return syn::Error::new_spanned(
            &impl_block,
            "#[command_handlers] must be on an inherent impl, not a trait impl",
        )
        .to_compile_error()
        .into();
    }

    let cmd_ty = &impl_block.self_ty;

    // Collect handler methods and extract state type
    let mut state_ty: Option<Type> = None;
    let mut handlers: Vec<HandlerMeta> = Vec::new();

    for item in &impl_block.items {
        let ImplItem::Fn(method) = item else { continue };

        // Skip non-async methods
        if method.sig.asyncness.is_none() {
            continue;
        }

        let method_name = &method.sig.ident;
        let variant_name = format_ident!("{}", method_name.to_string().to_pascal_case());

        // Extract params: expect (state: &T, ctx: Ctx, ...fields)
        let mut inputs = method.sig.inputs.iter();

        // First param: state: &SomePlugin
        let Some(FnArg::Typed(state_pat)) = inputs.next() else {
            return syn::Error::new_spanned(
                &method.sig,
                "handler must have at least (state: &Plugin, ctx: Ctx)",
            )
            .to_compile_error()
            .into();
        };

        // Extract and validate state type
        let extracted_state_ty = (*state_pat.ty).clone();
        match &state_ty {
            None => state_ty = Some(extracted_state_ty),
            Some(prev) => {
                // All handlers must use the same state type
                if quote!(#prev).to_string() != quote!(#extracted_state_ty).to_string() {
                    return syn::Error::new_spanned(
                        &state_pat.ty,
                        "all handlers must use the same state type",
                    )
                    .to_compile_error()
                    .into();
                }
            }
        }

        // Second param: ctx: Ctx (skip validation for now)
        let _ = inputs.next();

        // Remaining params are the variant fields
        let field_idents: Vec<Ident> = inputs
            .filter_map(|arg| {
                if let FnArg::Typed(PatType { pat, .. }) = arg {
                    if let Pat::Ident(pat_ident) = &**pat {
                        return Some(pat_ident.ident.clone());
                    }
                }
                None
            })
            .collect();

        handlers.push(HandlerMeta {
            method_name: method_name.clone(),
            variant_name,
            field_idents,
        });
    }

    let Some(state_ty) = state_ty else {
        return syn::Error::new_spanned(&impl_block, "no handler methods found")
            .to_compile_error()
            .into();
    };

    // Generate match arms
    let match_arms = handlers.iter().map(|h| {
        let method_name = &h.method_name;
        let variant_name = &h.variant_name;
        let fields = &h.field_idents;

        if fields.is_empty() {
            quote! {
                Self::#variant_name => Self::#method_name(state, ctx).await,
            }
        } else {
            quote! {
                Self::#variant_name { #( #fields ),* } => Self::#method_name(state, ctx, #( #fields ),*).await,
            }
        }
    });

    let execute_impl = quote! {
        impl #cmd_ty {
            pub async fn __execute(self, state: #state_ty, ctx: dragonfly_plugin::command::Ctx<'_>) {
                match self {
                    #( #match_arms )*
                }
            }
        }
    };

    let original = quote! { #impl_block };

    quote! {
        #original
        #execute_impl
    }
    .into()
}

struct HandlerMeta {
    method_name: Ident,
    variant_name: Ident,
    field_idents: Vec<Ident>,
}
