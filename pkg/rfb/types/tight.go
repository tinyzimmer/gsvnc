package types

// TightCapability represents a TightSecurity capability.
type TightCapability struct {
	Code              int32
	Vendor, Signature string
}
