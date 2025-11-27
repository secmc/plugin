//! Shared AST helpers used by the Rust SDK `xtask` generators.
//!
//! This module wraps common syn-based utilities for:
//! - Formatting generated Rust code (`prettify_code`).
//! - Scanning the prost-generated AST for structs and nested enums.
//! - Mapping prost field/attribute shapes into public-facing SDK types.
//! - Deciding how to convert helper method arguments into wire types.
//!
//! The functions here are internal to `xtask`, but the behavior they encode
//! (e.g. how `GameMode` enums or `Option<T>` fields are surfaced) is part of
//! the stable shape of the generated SDK API.

use anyhow::Result;
use quote::quote;
use std::collections::HashMap;
use syn::{
    parse::Parse, parse_file, visit::Visit, Field, File, GenericArgument, Ident, Item, ItemEnum,
    ItemMod, ItemStruct, Meta, Path, PathArguments, Type, TypePath,
};

/// Pretty-print a Rust source string using `prettyplease`.
///
/// This is applied to all generated files to keep diffs readable.
pub(crate) fn prettify_code(content: String) -> Result<String> {
    let ast = parse_file(&content)?;
    Ok(prettyplease::unparse(&ast))
}

pub(crate) fn find_all_structs(ast: &File) -> HashMap<Ident, &ItemStruct> {
    struct StructVisitor<'a> {
        structs: HashMap<Ident, &'a ItemStruct>,
    }
    impl<'a> Visit<'a> for StructVisitor<'a> {
        fn visit_item_struct(&mut self, i: &'a ItemStruct) {
            self.structs.insert(i.ident.clone(), i);
        }
        fn visit_item_mod(&mut self, i: &'a ItemMod) {
            if let Some((_, items)) = &i.content {
                for item in items {
                    self.visit_item(item);
                }
            }
        }
    }
    let mut visitor = StructVisitor {
        structs: Default::default(),
    };
    visitor.visit_file(ast);
    visitor.structs
}

pub(crate) fn find_nested_enum<'a>(
    ast: &'a File,
    mod_name: &str,
    enum_name: &str,
) -> Result<&'a ItemEnum> {
    for item in &ast.items {
        if let Item::Mod(item_mod) = item {
            if item_mod.ident == mod_name {
                if let Some((_, items)) = &item_mod.content {
                    for item in items {
                        if let Item::Enum(item_enum) = item {
                            if item_enum.ident == enum_name {
                                return Ok(item_enum);
                            }
                        }
                    }
                }
            }
        }
    }
    anyhow::bail!("Enum `{}::{}` not found in AST", mod_name, enum_name)
}

pub(crate) fn get_variant_type_path(variant: &syn::Variant) -> Result<&Path> {
    if let syn::Fields::Unnamed(fields) = &variant.fields {
        if let Some(field) = fields.unnamed.first() {
            if let Type::Path(type_path) = &field.ty {
                return Ok(&type_path.path);
            }
        }
    }
    anyhow::bail!("Variant `{}` is not a single-tuple struct", variant.ident)
}

/// Helper for clean_type to detect `Vec<u8>` variants by parsing
/// the syn::TypePath instead of string-matching.
fn is_vec_u8(type_path: &TypePath) -> bool {
    // Check if the last segment's identifier is "Vec"
    if let Some(segment) = type_path.path.segments.last() {
        if segment.ident == "Vec" {
            // Check if its type argument is `<u8>`
            if let syn::PathArguments::AngleBracketed(args) = &segment.arguments {
                if args.args.len() == 1 {
                    // Check for `Vec<u8>`
                    if let Some(syn::GenericArgument::Type(Type::Path(inner_type_path))) =
                        args.args.first()
                    {
                        // `is_ident` is robust for primitives
                        if inner_type_path.path.is_ident("u8") {
                            return true;
                        }
                    }
                }
            }
        }
    }
    false
}

enum ProstTypeInfo<'a> {
    Option(&'a Type),
    Vec(&'a Type),
    VecU8,
    HashMap(&'a Type, &'a Type),
    String,
    Primitive(&'a Ident),
    #[allow(dead_code)]
    Model(&'a Ident), // Any other struct/enum like Vec3, ItemStack
    Unknown(&'a Type),
}

fn classify_prost_type(ty: &Type) -> ProstTypeInfo<'_> {
    if let Type::Path(type_path) = ty {
        let segment = type_path.path.segments.last().unwrap();
        let ident = &segment.ident;
        let ident_str = ident.to_string();

        if ident_str == "Option" {
            if let PathArguments::AngleBracketed(args) = &segment.arguments {
                if let Some(GenericArgument::Type(inner_ty)) = args.args.first() {
                    return ProstTypeInfo::Option(inner_ty);
                }
            }
        }
        if ident_str == "Vec" {
            if is_vec_u8(type_path) {
                return ProstTypeInfo::VecU8;
            }
            if let PathArguments::AngleBracketed(args) = &segment.arguments {
                if let Some(GenericArgument::Type(inner_ty)) = args.args.first() {
                    return ProstTypeInfo::Vec(inner_ty);
                }
            }
        }

        if ident_str == "HashMap" {
            if let PathArguments::AngleBracketed(args) = &segment.arguments {
                if let (Some(GenericArgument::Type(k_ty)), Some(GenericArgument::Type(v_ty))) =
                    (args.args.first(), args.args.last())
                {
                    return ProstTypeInfo::HashMap(k_ty, v_ty);
                }
            }
        }

        if ident_str == "String" {
            return ProstTypeInfo::String;
        }
        if ["i32", "i64", "u32", "u64", "f32", "f64", "bool"].contains(&ident_str.as_str()) {
            return ProstTypeInfo::Primitive(ident);
        }

        return ProstTypeInfo::Model(ident);
    }

    ProstTypeInfo::Unknown(ty)
}

pub(crate) fn get_prost_enumeration(field: &Field) -> Option<Ident> {
    for attr in &field.attrs {
        // We only care about `#[prost(...)]` attributes
        if !attr.path().is_ident("prost") {
            continue;
        }

        // We expect the attribute to be a list, like `prost(string, tag="1")`
        if let Ok(list) = attr.meta.require_list() {
            // We need to parse the tokens inside the `(...)`
            // This is complex, so we parse them as a series of Meta items
            let result = list.parse_args_with(|input: syn::parse::ParseStream| {
                input.parse_terminated(Meta::parse, syn::Token![,])
            });

            if let Ok(parsed_metas) = result {
                for meta in parsed_metas {
                    // We are looking for a `NameValue` pair
                    if let Meta::NameValue(name_value) = meta {
                        // Specifically, one where the name is `enumeration`
                        if name_value.path.is_ident("enumeration") {
                            // The value should be a string literal, e.g., "GameMode"
                            if let syn::Expr::Lit(expr_lit) = name_value.value {
                                if let syn::Lit::Str(lit_str) = expr_lit.lit {
                                    return Some(lit_str.parse().unwrap());
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    // No `enumeration` attribute found
    None
}

pub(crate) fn get_api_type(field: &Field) -> proc_macro2::TokenStream {
    let enum_ident = get_prost_enumeration(field);

    match (enum_ident, classify_prost_type(&field.ty)) {
        // Enum field (prost represents as i32)
        (Some(ident), ProstTypeInfo::Primitive(_)) => {
            quote! { types::#ident }
        }
        // Option<Enum> field (prost represents as Option<i32>)
        (Some(ident), ProstTypeInfo::Option(_)) => {
            quote! { impl Into<Option<types::#ident>> }
        }
        // Enum with unexpected wrapper - just use the enum type
        (Some(ident), _) => {
            quote! { types::#ident }
        }
        // Not an enum, resolve normally
        (None, _) => resolve_recursive_api_type(&field.ty),
    }
}

/// Recursive helper for get_api_type.
/// Maps a prost type to its "clean" public-facing type.
fn resolve_recursive_api_type(ty: &Type) -> proc_macro2::TokenStream {
    match classify_prost_type(ty) {
        ProstTypeInfo::Option(inner) => {
            let inner_ty = resolve_recursive_api_type(inner); // Recurse
            quote! { impl Into<Option<#inner_ty>> }
        }
        ProstTypeInfo::Vec(inner) => {
            let inner_ty = resolve_recursive_api_type(inner); // Recurse
            quote! { Vec<#inner_ty> }
        }
        ProstTypeInfo::HashMap(k, v) => {
            let k_ty = resolve_recursive_api_type(k); // Recurse
            let v_ty = resolve_recursive_api_type(v); // Recurse
            quote! { std::collections::HashMap<#k_ty, #v_ty> }
        }
        ProstTypeInfo::VecU8 => quote! { Vec<u8> },
        ProstTypeInfo::String => quote! { String },
        ProstTypeInfo::Primitive(ident) => quote! { #ident },

        ProstTypeInfo::Model(_) => quote! { types::#ty },
        ProstTypeInfo::Unknown(ty) => quote! { #ty },
    }
}

pub(crate) enum ConversionLogic {
    /// No conversion needed, just use the name.
    Direct,
    /// Needs `.into()`.
    Into,
    /// Option<Enum>: needs `.into().map(|x| x.into())`
    OptionInnerInto,
}

pub(crate) fn get_action_conversion_logic(field: &Field) -> ConversionLogic {
    let is_enum = get_prost_enumeration(field).is_some();
    let is_option = matches!(classify_prost_type(&field.ty), ProstTypeInfo::Option(_));

    match (is_enum, is_option) {
        // Option<Enum> needs: .into().map(|x| x.into())
        (true, true) => ConversionLogic::OptionInnerInto,
        // Plain enum or plain Option needs: .into()
        (true, false) | (false, true) => ConversionLogic::Into,
        // Everything else: direct
        (false, false) => ConversionLogic::Direct,
    }
}

#[cfg(test)]
mod tests {
    use super::*; // Import all functions from the parent module
    use anyhow::Result;
    use quote::{format_ident, quote};
    use syn::{parse_file, parse_str, File, ItemEnum, Type};

    // --- Helper: Parse a string into a syn::File ---
    fn parse_test_file(code: &str) -> File {
        parse_file(code).expect("Failed to parse test code")
    }

    // --- Helper: Parse a string into a syn::Type ---
    fn parse_test_type(type_str: &str) -> Type {
        parse_str(type_str).expect("Failed to parse test type")
    }

    // --- Helper: Get realistic &Field objects from a mock struct ---
    fn get_test_fields() -> HashMap<String, Field> {
        let code = r#"
            struct MockAction {
                #[prost(string, tag = "1")]
                player_uuid: String,

                #[prost(enumeration = "GameMode", tag = "2")]
                game_mode: i32,

                #[prost(message, optional, tag = "3")]
                item: ::core::option::Option<ItemStack>,

                #[prost(message, tag = "4")]
                position: types::Vec3,
            }
        "#;
        let ast = parse_test_file(code);
        let structs = find_all_structs(&ast);
        let mock_struct = structs
            .get(&format_ident!("MockAction"))
            .expect("MockAction struct not found");

        mock_struct
            .fields
            .iter()
            .map(|f| (f.ident.as_ref().unwrap().to_string(), f.clone()))
            .collect()
    }

    // --- Tests for find_all_structs ---
    #[test]
    fn test_find_all_structs_nested() {
        let code = r#"
            struct TopLevel {}
            mod my_mod {
                struct NestedStruct {}
            }
        "#;
        let ast = parse_test_file(code);
        let structs = find_all_structs(&ast);
        assert_eq!(structs.len(), 2);
        assert!(structs.contains_key(&format_ident!("TopLevel")));
        assert!(structs.contains_key(&format_ident!("NestedStruct")));
    }

    // --- Tests for find_nested_enum ---
    #[test]
    fn test_find_nested_enum_success() {
        let code = r#"
            mod event_envelope {
                enum Payload { PlayerJoin(String) }
            }
        "#;
        let ast = parse_test_file(code);
        let result = find_nested_enum(&ast, "event_envelope", "Payload");
        assert!(result.is_ok());
    }

    #[test]
    fn test_find_nested_enum_fail() {
        let code = r#"
            mod event_envelope {
                enum Payload {}
            }
        "#;
        let ast = parse_test_file(code);
        let result = find_nested_enum(&ast, "wrong_mod", "Payload");
        assert!(result.is_err());
    }

    // --- Tests for get_variant_type_path ---
    #[test]
    fn test_get_variant_type_path() -> Result<()> {
        let item_enum: ItemEnum = parse_str(r#"enum E { Simple(String), Empty, Struct{f: i32} }"#)?;
        let simple_variant = item_enum.variants.first().unwrap();
        assert_eq!(quote!(#simple_variant).to_string(), "Simple (String)");
        let path = get_variant_type_path(simple_variant)?;
        assert_eq!(quote!(#path).to_string(), "String");

        let empty_variant = &item_enum.variants[1];
        assert!(get_variant_type_path(empty_variant).is_err());

        let struct_variant = &item_enum.variants[2];
        assert!(get_variant_type_path(struct_variant).is_err());
        Ok(())
    }

    // --- Tests for is_vec_u8 ---
    // first time ive ever had to actually use ref p some real ball knowledge here.
    // AI didn't even get it right
    #[test]
    fn test_is_vec_u8_helper() {
        let ty_vec_u8 = parse_test_type("Vec<u8>");
        let ty_prost_vec = parse_test_type("::prost::alloc::vec::Vec<u8>");
        let ty_vec_string = parse_test_type("Vec<String>");
        assert!(is_vec_u8(match ty_vec_u8 {
            Type::Path(ref p) => p,
            _ => panic!(),
        }));
        assert!(is_vec_u8(match ty_prost_vec {
            Type::Path(ref p) => p,
            _ => panic!(),
        }));
        assert!(!is_vec_u8(match ty_vec_string {
            Type::Path(ref p) => p,
            _ => panic!(),
        }));
    }

    // --- Tests for classify_prost_type ---
    #[test]
    fn test_classify_prost_type() {
        let ty_opt = parse_test_type("Option<types::Vec3>");
        let ty_vec = parse_test_type("Vec<String>");
        let ty_bytes = parse_test_type("::prost::alloc::vec::Vec<u8>");
        let ty_map = parse_test_type("::std::collections::HashMap<String, i32>");
        let ty_str = parse_test_type("::prost::alloc::string::String");
        let ty_prim = parse_test_type("i32");
        let ty_model = parse_test_type("types::ItemStack");

        assert!(matches!(
            classify_prost_type(&ty_opt),
            ProstTypeInfo::Option(_)
        ));
        assert!(matches!(
            classify_prost_type(&ty_vec),
            ProstTypeInfo::Vec(_)
        ));
        assert!(matches!(
            classify_prost_type(&ty_bytes),
            ProstTypeInfo::VecU8
        ));
        assert!(matches!(
            classify_prost_type(&ty_map),
            ProstTypeInfo::HashMap(_, _)
        ));
        assert!(matches!(
            classify_prost_type(&ty_str),
            ProstTypeInfo::String
        ));
        assert!(matches!(
            classify_prost_type(&ty_prim),
            ProstTypeInfo::Primitive(_)
        ));
        assert!(matches!(
            classify_prost_type(&ty_model),
            ProstTypeInfo::Model(_)
        ));
    }

    // --- Tests for get_prost_enumeration ---
    #[test]
    fn test_get_prost_enumeration() {
        let fields = get_test_fields();
        let field_game_mode = fields.get("game_mode").unwrap();
        let field_uuid = fields.get("player_uuid").unwrap();
        let field_item = fields.get("item").unwrap();

        assert_eq!(
            get_prost_enumeration(field_game_mode).unwrap().to_string(),
            "GameMode"
        );
        assert!(get_prost_enumeration(field_uuid).is_none());
        assert!(get_prost_enumeration(field_item).is_none());
    }

    // --- Tests for resolve_recursive_api_type ---
    fn assert_resolve_api_type(input_type: &str, expected_output: &str) {
        let ty = parse_test_type(input_type);
        let resolved = resolve_recursive_api_type(&ty);
        assert_eq!(
            resolved.to_string().replace(' ', ""),
            expected_output.replace(' ', "")
        );
    }

    #[test]
    fn test_resolve_recursive_api_type() {
        assert_resolve_api_type("i32", "i32");
        assert_resolve_api_type("::prost::alloc::string::String", "String");
        assert_resolve_api_type("::prost::alloc::vec::Vec<u8>", "Vec<u8>");
        assert_resolve_api_type("ItemStack", "types::ItemStack");
        assert_resolve_api_type("Option<Vec3>", "implInto<Option<types::Vec3>>");
        assert_resolve_api_type("Vec<ItemStack>", "Vec<types::ItemStack>");
        assert_resolve_api_type(
            "std::collections::HashMap<String, BlockState>",
            "std::collections::HashMap<String,types::BlockState>",
        );
    }

    // --- Tests for get_api_type ---
    fn assert_api_type(field: &Field, expected_output: &str) {
        let resolved = get_api_type(field);
        assert_eq!(
            resolved.to_string().replace(' ', ""),
            expected_output.replace(' ', "")
        );
    }

    #[test]
    fn test_get_api_type() {
        let fields = get_test_fields();

        // Test enum field
        assert_api_type(fields.get("game_mode").unwrap(), "types::GameMode");

        // Test string field
        assert_api_type(fields.get("player_uuid").unwrap(), "String");

        // Test optional model field
        assert_api_type(
            fields.get("item").unwrap(),
            "implInto<Option<types::ItemStack>>",
        );
    }

    // --- Tests for get_action_conversion_logic ---
    #[test]
    fn test_get_action_conversion_logic() {
        let fields = get_test_fields();
        let field_game_mode = fields.get("game_mode").unwrap();
        let field_uuid = fields.get("player_uuid").unwrap();
        let field_item = fields.get("item").unwrap();
        let field_pos = fields.get("position").unwrap();

        // Enum -> Into
        assert!(matches!(
            get_action_conversion_logic(field_game_mode),
            ConversionLogic::Into
        ));

        // Option -> Into
        assert!(matches!(
            get_action_conversion_logic(field_item),
            ConversionLogic::Into
        ));

        // String -> Direct
        assert!(matches!(
            get_action_conversion_logic(field_uuid),
            ConversionLogic::Direct
        ));

        // Model (types::Vec3) -> Direct
        assert!(matches!(
            get_action_conversion_logic(field_pos),
            ConversionLogic::Direct
        ));
    }
}
