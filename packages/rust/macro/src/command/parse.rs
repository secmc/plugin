use syn::{
    parse::{Parse, ParseStream},
    punctuated::Punctuated,
    spanned::Spanned,
    Attribute, Expr, ExprLit, Lit, LitStr, Meta, Token,
};

pub(crate) struct CommandInfoParser {
    pub name: LitStr,
    pub description: LitStr,
    pub aliases: Vec<LitStr>,
}

impl Parse for CommandInfoParser {
    fn parse(input: ParseStream) -> syn::Result<Self> {
        let metas = Punctuated::<Meta, Token![,]>::parse_terminated(input)?;

        let mut name = None;
        let mut description = None;
        let mut aliases = Vec::new();

        for meta in metas {
            match meta {
                Meta::NameValue(nv) if nv.path.is_ident("name") => {
                    if let Expr::Lit(ExprLit {
                        lit: Lit::Str(s), ..
                    }) = nv.value
                    {
                        name = Some(s);
                    } else {
                        return Err(syn::Error::new(
                            nv.value.span(),
                            "expected string literal for `name`",
                        ));
                    }
                }
                Meta::NameValue(nv) if nv.path.is_ident("description") => {
                    if let Expr::Lit(ExprLit {
                        lit: Lit::Str(s), ..
                    }) = nv.value
                    {
                        description = Some(s);
                    } else {
                        return Err(syn::Error::new(
                            nv.value.span(),
                            "expected string literal for `description`",
                        ));
                    }
                }
                Meta::List(list) if list.path.is_ident("aliases") => {
                    aliases = list
                        .parse_args_with(Punctuated::<LitStr, Token![,]>::parse_terminated)?
                        .into_iter()
                        .collect();
                }
                _ => {
                    return Err(syn::Error::new(
                        meta.span(),
                        "unrecognized command attribute",
                    ));
                }
            }
        }

        Ok(Self {
            name: name.ok_or_else(|| {
                syn::Error::new(input.span(), "missing required attribute `name`")
            })?,
            description: description.unwrap_or_else(|| LitStr::new("", input.span())),
            aliases,
        })
    }
}

#[derive(Default)]
pub(crate) struct SubcommandAttr {
    pub name: Option<LitStr>,
    pub aliases: Vec<LitStr>,
}

pub(crate) fn parse_subcommand_attr(attrs: &[Attribute]) -> syn::Result<SubcommandAttr> {
    let mut out = SubcommandAttr::default();

    for attr in attrs {
        if !attr.path().is_ident("subcommand") {
            continue;
        }

        let metas = attr.parse_args_with(Punctuated::<Meta, Token![,]>::parse_terminated)?;

        for meta in metas {
            match meta {
                // name = "pay"
                Meta::NameValue(nv) if nv.path.is_ident("name") => {
                    if let Expr::Lit(ExprLit {
                        lit: Lit::Str(s), ..
                    }) = nv.value
                    {
                        out.name = Some(s);
                    } else {
                        return Err(syn::Error::new_spanned(
                            nv.value,
                            "subcommand `name` must be a string literal",
                        ));
                    }
                }

                // aliases("give", "send")
                Meta::List(list) if list.path.is_ident("aliases") => {
                    out.aliases = list
                        .parse_args_with(Punctuated::<LitStr, Token![,]>::parse_terminated)?
                        .into_iter()
                        .collect();
                }

                _ => {
                    return Err(syn::Error::new_spanned(
                        meta,
                        "unknown subcommand attribute; expected `name = \"...\"` or `aliases(..)`",
                    ));
                }
            }
        }
    }

    Ok(out)
}


