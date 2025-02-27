package common

// PromiseOrigin is where the promise come from.
type PromiseOrigin int

const (
	// PromiseOriginRef is when a promise object is created from a ref in document.
	PromiseOriginRef PromiseOrigin = iota
	// PromiseOriginInternal is when a promise object is created internally for internal use only.
	PromiseOriginInternal
)

// ObjectPromise is the object used for late-binding between the compiled artifacts. Since the artifacts are compiled
// independently, when we need to another artifact that semantically bound with the one, we should use a promise object.
//
// For example, when we compile a channel and want to use a server that it point to, we create a promise object to
// the server, and it becomes available after the linking stage.
//
// Every promise can contain a ref where the object is located in document (or external document) or a callback function
// to find the object in the storage. The functions Ref and FindCallback provide that. Once an object is found, the
// linker calls the Assign method to bind the object to the promise.
type ObjectPromise interface {
	// Assign binds the object to the promise. Called by the linker.
	Assign(obj Artifact)
	// Assigned returns true if the object is already bound to the promise.
	Assigned() bool
	// Ref returns the reference to the object if any.
	Ref() string
	Origin() PromiseOrigin
	// FindCallback returns the callback function to find the object in the storage. If not nil, Ref is ignored.
	//
	// The linker calls this function for every artifact in all storages to find the one that matches this promise.
	// If the callback returns true, the linker binds the object to the promise. If it returns true more than once, the
	// linker aborts with error.
	FindCallback() PromiseFindCbFunc
}

// ObjectListPromise is the promise object like ObjectPromise but for the list of objects. It can't be referenced by
// ref and intended only for internal use.
type ObjectListPromise interface {
	// AssignList binds the list of objects to the promise. Called by the linker.
	AssignList(objs []Artifact)
	// Assigned returns true if the list of objects is already bound to the promise.
	Assigned() bool
	// FindCallback returns the callback function to find the object in the storage.
	//
	// The linker calls this function for every artifact in all storages to find the ones that match this promise.
	// If the callback returns true, the linker adds the object to the list of the promise. Can return true multiple times.
	FindCallback() PromiseFindCbFunc
}

type PromiseFindCbFunc func(item Artifact) bool
