package run

import "errors"

// ErrUnsealEnvelope is returned from Subscribe methods of channels or operations, indicating that the Envelope
// could not be unsealed (unmarshalled) into the message type.
var ErrUnsealEnvelope = errors.New("unseal envelope error")
