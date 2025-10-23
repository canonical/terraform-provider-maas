// Package testutils
package testutils

import (
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
)

// checkSemverConstraint was adapted from [terraform-provider-grafana](https://github.com/grafana/terraform-provider-grafana)
// Licensed under MPL-2.0
func checkSemverConstraint(t *testing.T, semverConstraint string) {
	t.Helper()

	versionStr := os.Getenv("MAAS_VERSION")
	if semverConstraint != "" && versionStr != "" {
		version := semver.MustParse(versionStr)

		c, err := semver.NewConstraint(semverConstraint)
		if err != nil {
			t.Fatalf("invalid constraint %s: %v", semverConstraint, err)
		}

		if !c.Check(version) {
			t.Skipf("skipping test for MAAS version `%s`, constraint `%s`", versionStr, semverConstraint)
		}
	}
}

// SkipTestIfNotMAASVersion skips acceptance tests if the MAAS version constraint is not met.
func SkipTestIfNotMAASVersion(t *testing.T, semverConstraint string) {
	t.Helper()

	if runAccTests := os.Getenv("TF_ACC"); runAccTests != "1" {
		return
	}

	checkSemverConstraint(t, semverConstraint)
}
