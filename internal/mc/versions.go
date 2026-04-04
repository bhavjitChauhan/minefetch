package mc

//go:generate go run gen_versions.go

// VersionIdName maps [protocol version numbers] (PVNs) to string representations.
//
// See the comment on VersionNameId for details on the string representation.
//
// For PVNs that map to multiple versions, only the earliest and latest versions are listed.
// For example, the PVN 47 maps to 22 versions, but only the range "1.8 – 1.8.9" is expressed.
//
// [protocol version numbers]: https://minecraft.wiki/w/Protocol_version
var VersionIdName map[int32]string = versionIdName

// VersionNameId maps version string representations to [protocol version numbers] (PVNs).
//
// The string representations are taken from the Minecraft Wiki [Protocol version script],
// with the following modifications:
//   - "Release Candidate X" becomes "-rcX"
//   - "Pre-Release X" becomes "-preX"
//	 - "latest" maps to the latest stable release, which is the largest PVN less than 800
//
// Multiple strings may map to the same PVN.
//
// [protocol version numbers]: https://minecraft.wiki/w/Protocol_version
// [Protocol version script]: https://minecraft.wiki/w/Module:Protocol_version/Versions
var VersionNameId map[string]int32 = versionNameId
