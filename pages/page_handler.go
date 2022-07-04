package pages

// Page handling proposes a hard question: How do you map
// virtual directories and pages? Of course, you could
// enforce a strict boundary between the file system
// and the file (see: FileHandler) but how would you
// otherwise allow routing dependent on:
// - the directory and any subdirectories
// - the final segment in the path, aka the 'file'
//
// You could view the directories as a kind of tree,
// where if the current traversed node is not a leaf,
// it will result in some kind of data fetch - and
// inevitably a response to the client. However,
// if the node is a branch, then it will result in
// a breadth first search (likely just a map val check)
// until the next node is found. Essentially, we're
// further registering handlers to this, composing our
// virtual set of directories and then traversing through
// them - once we hit a leaf, this is when we call the
// leaf handler, which can be many things. Examples:
// - virtual directory which checks if a page exists
// - endpoint, leading us to a single page in the tree
//
// These directories do not need to be near each other
// in the disk. They simply have to be properly registered
// in the tree in order to be read. Pages can even live in
// the program's memory, if required. (pages possibly don't
// even need to be *in* the filesystem, necessarily)
//
// This is different from FileHandler, as FileHandler maps
// a disk directory directly into the path. PageHandler instead
// maps virtual directories into nodes, further executing
// any function tied to the node.
//
// This should be generic enough to handle most requests. Anything
// that requires special processing should probably instead
// create its own handler.
