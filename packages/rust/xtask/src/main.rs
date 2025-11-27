//! Code generation entrypoint for the Rust `dragonfly-plugin` SDK.
//!
//! This `xtask` binary:
//! - Parses the prost-generated `src/generated/df.plugin.rs` file.
//! - Builds an in-memory AST of all events, actions, and result types.
//! - Regenerates three helper files in the main SDK crate:
//!   - `src/event/handler.rs` (`generate_handlers`)
//!   - `src/event/mutations.rs` (`generate_mutations`)
//!   - `src/server/helpers.rs` (`generate_actions`)
//! The public API of the SDK is considered stable; this task should only
//! be changed in ways that preserve the shape of the generated code.

pub mod generate_actions;
pub mod generate_handlers;
pub mod generate_mutations;
pub mod utils;

use anyhow::Result;
use std::{fs, path::PathBuf};
use syn::parse_file;

use crate::utils::find_all_structs;

fn main() -> Result<()> {
    println!("Starting 'xtask' code generation...");
    let root_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .parent()
        .unwrap()
        .to_path_buf();

    let generated_file_path = root_dir.join("src/generated/df.plugin.rs");
    let handler_output_path = root_dir.join("src/event/handler.rs");
    let mutations_output_path = root_dir.join("src/event/mutations.rs");
    let server_output_path = root_dir.join("src/server/helpers.rs");

    // only read the generated code file once we can just reference it out so we don't waste resources.
    println!("Parsing generated proto file...");
    let generated_code = fs::read_to_string(generated_file_path)?;
    let ast = parse_file(&generated_code)?;

    // just cache all structs, we can hashmap look them up em.
    let all_structs = find_all_structs(&ast);

    // first generate event handlers:
    generate_handlers::generate_handler_trait(&ast, &handler_output_path)?;

    generate_mutations::generate_event_mutations(&ast, &all_structs, &mutations_output_path)?;

    generate_actions::generate_server_helpers(&ast, &all_structs, &server_output_path)?;
    Ok(())
}
