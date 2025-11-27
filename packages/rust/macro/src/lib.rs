//! Procedural macros for the `dragonfly-plugin` Rust SDK.
//!
//! This crate exposes three main macros:
//! - `#[derive(Plugin)]` with a `#[plugin(...)]` attribute to describe
//!   the plugin metadata and registered commands.
//! - `#[event_handler]` to generate an `EventSubscriptions` implementation
//!   based on the `on_*` methods you override in an `impl EventHandler`.
//! - `#[derive(Command)]` together with `#[command(...)]` /
//!   `#[subcommand(...)]` to generate strongly-typed command parsers and
//!   handler traits.
//!
//! These macros are re-exported by the `dragonfly-plugin` crate, so plugin
//! authors should depend on that crate directly.

mod command;
mod plugin;

use heck::ToPascalCase;
use proc_macro::TokenStream;
use quote::{format_ident, quote};
use syn::{parse_macro_input, Attribute, DeriveInput, ImplItem, ItemImpl};

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

#[proc_macro_derive(Command, attributes(command, subcommand))]
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
