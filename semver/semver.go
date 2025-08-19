package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

type Constraint struct {
	Operator string
	Version  Version
}

var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`)


func ParseVersion(version string) (Version, error) {
	matches := semverRegex.FindStringSubmatch(version)
	if matches == nil {
		return Version{}, fmt.Errorf("invalid semantic version: %s", version)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}, nil
}


func (v Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		version += "-" + v.Prerelease
	}
	if v.Build != "" {
		version += "+" + v.Build
	}
	return version
}



func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1 
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1 
	}
	if v.Prerelease != other.Prerelease {
		if v.Prerelease < other.Prerelease {
			return -1
		}
		return 1
	}

	return 0
}


func ParseConstraint(constraint string) (Constraint, error) {
	constraint = strings.TrimSpace(constraint)
	
	if strings.HasPrefix(constraint, "^") {
		version, err := ParseVersion(constraint[1:])
		return Constraint{Operator: "^", Version: version}, err
	}
	
	if strings.HasPrefix(constraint, "~") {
		version, err := ParseVersion(constraint[1:])
		return Constraint{Operator: "~", Version: version}, err
	}
	
	if strings.HasPrefix(constraint, ">=") {
		version, err := ParseVersion(constraint[2:])
		return Constraint{Operator: ">=", Version: version}, err
	}
	
	if strings.HasPrefix(constraint, "<=") {
		version, err := ParseVersion(constraint[2:])
		return Constraint{Operator: "<=", Version: version}, err
	}
	
	if strings.HasPrefix(constraint, ">") {
		version, err := ParseVersion(constraint[1:])
		return Constraint{Operator: ">", Version: version}, err
	}
	
	if strings.HasPrefix(constraint, "<") {
		version, err := ParseVersion(constraint[1:])
		return Constraint{Operator: "<", Version: version}, err
	}
	
	if strings.HasPrefix(constraint, "=") {
		version, err := ParseVersion(constraint[1:])
		return Constraint{Operator: "=", Version: version}, err
	}
	
	
	version, err := ParseVersion(constraint)
	return Constraint{Operator: "=", Version: version}, err
}


func (c Constraint) Satisfies(version Version) bool {
	switch c.Operator {
	case "^":
		
		if version.Major != c.Version.Major {
			return false
		}
		return version.Compare(c.Version) >= 0
		
	case "~":
		
		if version.Major != c.Version.Major || version.Minor != c.Version.Minor {
			return false
		}
		return version.Compare(c.Version) >= 0
		
	case ">=":
		return version.Compare(c.Version) >= 0
		
	case "<=":
		return version.Compare(c.Version) <= 0
		
	case ">":
		return version.Compare(c.Version) > 0
		
	case "<":
		return version.Compare(c.Version) < 0
		
	case "=":
		return version.Compare(c.Version) == 0
		
	default:
		return false
	}
}


func FindBestMatch(constraint Constraint, versions []Version) (*Version, error) {
	var candidates []Version
	
	for _, version := range versions {
		if constraint.Satisfies(version) {
			candidates = append(candidates, version)
		}
	}
	
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no version satisfies constraint %s%s", constraint.Operator, constraint.Version)
	}
	
	
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.Compare(best) > 0 {
			best = candidate
		}
	}
	
	return &best, nil
}
