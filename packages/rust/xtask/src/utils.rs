use anyhow::Result;
use quote::quote;
use std::{collections::HashMap, fs, path::PathBuf};
use syn::{
    parse_file, visit::Visit, File, Ident, Item, ItemEnum, ItemMod, ItemStruct, Path, Type,
    TypePath,
};

pub(crate) fn write_formatted_file(path: &PathBuf, content: String) -> Result<()> {
    let ast = parse_file(&content)?;
    let formatted_code = prettyplease::unparse(&ast);
    fs::write(path, formatted_code)?;
    Ok(())
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

pub(crate) fn clean_type(ty: &Type) -> proc_macro2::TokenStream {
    let (inner_type, is_option) = unwrap_option_path(ty);
    if is_option {
        let inner_cleaned = clean_type(inner_type);
        return quote! { Option<#inner_cleaned> };
    }

    if let Type::Path(type_path) = inner_type {
        if let Some(segment) = type_path.path.segments.last() {
            let ident_str = segment.ident.to_string();

            match ident_str.as_str() {
                "String" => return quote! { String },

                "StringList" => return quote! { Vec<String> },
                "ItemStackList" => return quote! { Vec<types::ItemStack> },

                "Vec" if is_prost_bytes(type_path) => return quote! { Vec<u8> },

                // Primitives (safe, no path)
                "f64" => return quote! { f64 },
                "f32" => return quote! { f32 },
                "i64" => return quote! { i64 },
                "u64" => return quote! { u64 },
                "i32" => return quote! { i32 },
                "u32" => return quote! { u32 },
                "bool" => return quote! { bool },

                // Fallback for complex types (Vec3, WorldRef, etc.)
                // This is now correct. If `ident_str` is `Vec3`, it
                // will correctly become `types::Vec3`.
                _ => {
                    let ident = &segment.ident;
                    return quote! { types::#ident };
                }
            }
        }
    }

    // Last resort (e.g., tuples, arrays)
    quote! { #ty }
}

/// Helper for clean_type to detect `prost` bytes (Vec<u8>)
fn is_prost_bytes(type_path: &TypePath) -> bool {
    let path_str = quote!(#type_path).to_string();
    path_str.contains("Vec < u8 > ")
}
/// Unwraps `Option<T>` or `::core::option::Option<T>` and returns `T`
pub(crate) fn unwrap_option_path(ty: &Type) -> (&Type, bool) {
    if let Type::Path(type_path) = ty {
        if let Some(segment) = type_path.path.segments.last() {
            if segment.ident == "Option" {
                if let syn::PathArguments::AngleBracketed(args) = &segment.arguments {
                    if let Some(syn::GenericArgument::Type(inner_ty)) = args.args.first() {
                        return (inner_ty, true);
                    }
                }
            }
        }
    }
    (ty, false)
}

pub(crate) fn generate_conversion_code(
    var_name: &proc_macro2::TokenStream,
    original_ty: &Type,
) -> proc_macro2::TokenStream {
    let (inner_type, is_option) = unwrap_option_path(original_ty);
    if is_option {
        let inner_var = quote! { v };
        let inner_conversion = generate_conversion_code(&inner_var, inner_type);
        return quote! { #var_name.map(|v| #inner_conversion) };
    }

    if let Type::Path(type_path) = inner_type {
        if let Some(segment) = type_path.path.segments.last() {
            let ident_str = segment.ident.to_string();

            if ident_str == "StringList" {
                return quote! {
                    types::StringList {
                        values: #var_name.into_iter().map(|s| s.into()).collect(),
                    }
                };
            }
            if ident_str == "ItemStackList" {
                return quote! {
                    types::ItemStackList {
                        items: #var_name.into_iter().map(|s| s.into()).collect(),
                    }
                };
            }
        }
    }

    // Default case: just use .into()
    quote! { #var_name.into() }
}
