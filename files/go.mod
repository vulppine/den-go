module files

go 1.18

// wow! I do not like this
// some kinda really odd non-code file
// to dictate how submodule dependencies
// should be imported
require den/routing v0.0.0
replace den/routing => ./../routing