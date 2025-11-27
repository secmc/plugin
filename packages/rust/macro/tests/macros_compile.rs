#[test]
fn macros_compile_on_basic_examples() {
    let t = trybuild::TestCases::new();
    t.pass("tests/fixtures/plugin_basic.rs");
    t.pass("tests/fixtures/command_basic.rs");
}


