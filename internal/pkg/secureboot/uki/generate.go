// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package uki

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/siderolabs/gen/xslices"

	"github.com/siderolabs/talos/internal/pkg/secureboot"
	"github.com/siderolabs/talos/internal/pkg/secureboot/measure"
	"github.com/siderolabs/talos/pkg/machinery/constants"
	"github.com/siderolabs/talos/pkg/splash"
	"github.com/siderolabs/talos/pkg/version"
)

func (builder *Builder) generateOSRel() error {
	osRelease, err := version.OSRelease()
	if err != nil {
		return err
	}

	path := filepath.Join(builder.scratchDir, "os-release")

	if err = os.WriteFile(path, osRelease, 0o600); err != nil {
		return err
	}

	builder.sections = append(builder.sections,
		section{
			Name:    secureboot.OSRel,
			Path:    path,
			Measure: true,
			Append:  true,
		},
	)

	return nil
}

func (builder *Builder) generateCmdline() error {
	path := filepath.Join(builder.scratchDir, "cmdline")

	if err := os.WriteFile(path, []byte(builder.Cmdline), 0o600); err != nil {
		return err
	}

	builder.sections = append(builder.sections,
		section{
			Name:    secureboot.CMDLine,
			Path:    path,
			Measure: true,
			Append:  true,
		},
	)

	return nil
}

func (builder *Builder) generateInitrd() error {
	builder.sections = append(builder.sections,
		section{
			Name:    secureboot.Initrd,
			Path:    builder.InitrdPath,
			Measure: true,
			Append:  true,
		},
	)

	return nil
}

func (builder *Builder) generateSplash() error {
	path := filepath.Join(builder.scratchDir, "splash.bmp")

	if err := os.WriteFile(path, splash.GetBootImage(), 0o600); err != nil {
		return err
	}

	builder.sections = append(builder.sections,
		section{
			Name:    secureboot.Splash,
			Path:    path,
			Measure: true,
			Append:  true,
		},
	)

	return nil
}

func (builder *Builder) generateUname() error {
	path := filepath.Join(builder.scratchDir, "uname")

	if err := os.WriteFile(path, []byte(constants.DefaultKernelVersion), 0o600); err != nil {
		return err
	}

	builder.sections = append(builder.sections,
		section{
			Name:    secureboot.Uname,
			Path:    path,
			Measure: true,
			Append:  true,
		},
	)

	return nil
}

func (builder *Builder) generateSBAT() error {
	sbat, err := GetSBAT(builder.SdStubPath)
	if err != nil {
		return err
	}

	path := filepath.Join(builder.scratchDir, "sbat")

	if err = os.WriteFile(path, sbat, 0o600); err != nil {
		return err
	}

	builder.sections = append(builder.sections,
		section{
			Name:    secureboot.SBAT,
			Path:    path,
			Measure: true,
		},
	)

	return nil
}

func (builder *Builder) generatePCRPublicKey() error {
	builder.sections = append(builder.sections,
		section{
			Name:    secureboot.PCRPKey,
			Path:    builder.PCRPublicKeyPath,
			Append:  true,
			Measure: true,
		},
	)

	return nil
}

func (builder *Builder) generateKernel() error {
	path := filepath.Join(builder.scratchDir, "kernel")

	if err := builder.peSigner.Sign(builder.KernelPath, path); err != nil {
		return err
	}

	builder.sections = append(builder.sections,
		section{
			Name:    secureboot.Linux,
			Path:    path,
			Append:  true,
			Measure: true,
		},
	)

	return nil
}

func (builder *Builder) generatePCRSig() error {
	sectionsData := xslices.ToMap(
		xslices.Filter(builder.sections,
			func(s section) bool {
				return s.Measure
			},
		),
		func(s section) (secureboot.Section, string) {
			return s.Name, s.Path
		})

	pcrData, err := measure.GenerateSignedPCR(sectionsData, builder.PCRSigningKeyPath)
	if err != nil {
		return err
	}

	pcrSignatureData, err := json.Marshal(pcrData)
	if err != nil {
		return err
	}

	path := filepath.Join(builder.scratchDir, "pcrpsig")

	if err = os.WriteFile(path, pcrSignatureData, 0o600); err != nil {
		return err
	}

	builder.sections = append(builder.sections,
		section{
			Name:   secureboot.PCRSig,
			Path:   path,
			Append: true,
		},
	)

	return nil
}
