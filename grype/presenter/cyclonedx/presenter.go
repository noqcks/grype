package cyclonedx

import (
	"io"

	"github.com/CycloneDX/cyclonedx-go"

	"github.com/anchore/grype/grype/match"
	"github.com/anchore/grype/grype/pkg"
	"github.com/anchore/grype/grype/presenter/models"
	"github.com/anchore/grype/grype/vulnerability"
	"github.com/anchore/grype/internal"
	"github.com/anchore/grype/internal/version"
	"github.com/anchore/syft/syft/formats/common/cyclonedxhelpers"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
)

// Presenter writes a CycloneDX report from the given Matches and Scope contents
type Presenter struct {
	results          match.Matches
	packages         []pkg.Package
	srcMetadata      *source.Metadata
	metadataProvider vulnerability.MetadataProvider
	format           cyclonedx.BOMFileFormat
	sbom             *sbom.SBOM
}

// NewPresenter is a *Presenter constructor
func NewJSONPresenter(pb models.PresenterConfig) *Presenter {
	return &Presenter{
		results:          pb.Matches,
		packages:         pb.Packages,
		metadataProvider: pb.MetadataProvider,
		srcMetadata:      pb.Context.Source,
		sbom:             pb.SBOM,
		format:           cyclonedx.BOMFileFormatJSON,
	}
}

// NewPresenter is a *Presenter constructor
func NewXMLPresenter(pb models.PresenterConfig) *Presenter {
	return &Presenter{
		results:          pb.Matches,
		packages:         pb.Packages,
		metadataProvider: pb.MetadataProvider,
		srcMetadata:      pb.Context.Source,
		sbom:             pb.SBOM,
		format:           cyclonedx.BOMFileFormatXML,
	}
}

// Present creates a CycloneDX-based reporting
func (pres *Presenter) Present(output io.Writer) error {
	// note: this uses the syft cyclondx helpers to create
	// a consistent cyclondx BOM across syft and grype
	cyclonedxBOM := cyclonedxhelpers.ToFormatModel(*pres.sbom)

	// empty the tool metadata and add grype metadata
	versionInfo := version.FromBuild()
	cyclonedxBOM.Metadata.Tools = &[]cyclonedx.Tool{
		{
			Vendor:  "anchore",
			Name:    internal.ApplicationName,
			Version: versionInfo.Version,
		},
	}

	vulns := make([]cyclonedx.Vulnerability, 0)
	for m := range pres.results.Enumerate() {
		v, err := NewVulnerability(m, pres.metadataProvider)
		if err != nil {
			continue
		}
		vulns = append(vulns, v)
	}
	cyclonedxBOM.Vulnerabilities = &vulns
	enc := cyclonedx.NewBOMEncoder(output, pres.format)
	enc.SetPretty(true)
	enc.SetEscapeHTML(false)

	return enc.Encode(cyclonedxBOM)
}
