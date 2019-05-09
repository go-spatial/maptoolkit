package atlante

import (
	"io"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/go-spatial/maptoolkit/atlante/grids"
	"github.com/go-spatial/maptoolkit/atlante/internal/urlutil"
)

var funcMap = template.FuncMap{
	"to_upper":    strings.ToUpper,
	"to_lower":    strings.ToLower,
	"format":      tplFormat,
	"now":         time.Now,
	"div":         tplMathDiv,
	"add":         tplMathAdd,
	"sub":         tplMathSub,
	"mul":         tplMathMul,
	"neg":         tplMathNeg,
	"abs":         tplMathAbs,
	"seq":         tplSeq,
	"new_toggler": tplNewToggle,
	"rounder_for": tplRoundTo,
	"rounder3":    tplRound3,
	"first":       tplFirstNonZero,
}

// Sheet describes a map sheet
type Sheet struct {
	Name string
	// Data provider for this sheet
	grids.Provider
	// DPI at which the sheet should be rendered. Defaults to 144
	DPI uint
	// Scale value (50000, 5000, etc...)
	Scale uint
	// URL to the style file
	Style string
	// Template file to use
	SvgTemplateFilename string

	// Parsed and ready template
	svgTemplate *template.Template
}

// NewSheet returns a new sheet
func NewSheet(name string, provider grids.Provider, dpi uint, scale uint, style string, svgTemplateFilename *url.URL) (*Sheet, error) {
	var (
		err error
		t   *template.Template
	)

	name = strings.TrimSpace(strings.ToLower(name))

	tpl, err := urlutil.ReadAll(svgTemplateFilename)
	if err != nil {
		return nil, err
	}

	t, err = template.New(svgTemplateFilename.String()).
		Funcs(funcMap).
		Option("missingkey=error").
		Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	return &Sheet{
		Name:                name,
		Provider:            provider,
		DPI:                 dpi,
		Scale:               scale,
		Style:               style,
		SvgTemplateFilename: svgTemplateFilename.String(),
		svgTemplate:         t,
	}, nil
}

// Execute the sheets template
func (sheet *Sheet) Execute(wr io.Writer, tplContext GridTemplateContext) error {
	return sheet.svgTemplate.Execute(wr, tplContext)
}

// SheetFor returns the sheet for the given name, if the sheet does not exists
// sheet.ErrUnkownSheetName is returned.
func (a *Atlante) SheetFor(sheetName string) (*Sheet, error) {
	if a == nil {
		return nil, ErrNilAtlanteObject
	}
	if len(a.sheets) == 0 {
		return nil, ErrNoSheets
	}
	sheetnm := strings.TrimSpace(strings.ToLower(sheetName))

	a.sLock.RLock()
	sheet := a.sheets[sheetnm]
	a.sLock.RUnlock()
	if sheet == nil {
		return nil, ErrUnknownSheetName(sheetnm)
	}
	return sheet, nil
}

// Sheets returns the currently configured sheet names.
func (a *Atlante) Sheets() (sheets []string) {
	if a == nil || len(a.sheets) == 0 {
		return sheets
	}
	sheets = make([]string, len(a.sheets))
	a.sLock.RLock()
	i := 0
	for k := range a.sheets {
		sheets[i] = k
		i++
	}
	a.sLock.RUnlock()
	return sheets
}

// AddSheet will add a sheet to the instance of atlante.
// Error that can be generated are `ErrBlankSheetName` and `ErrDuplicateSheetName`
// The name of the sheet is normalize to lowercase and spaces trimmed
func (a *Atlante) AddSheet(s *Sheet) error {
	if s == nil {
		return ErrNilSheet
	}
	name := strings.TrimSpace(strings.ToLower(s.Name))
	if name == "" {
		return ErrBlankSheetName
	}
	a.sLock.Lock()
	defer a.sLock.Unlock()
	if a.sheets == nil {
		a.sheets = make(map[string]*Sheet)
		a.sheets[name] = s
		return nil
	}
	if _, ok := a.sheets[name]; ok {
		return ErrDuplicateSheetName
	}
	a.sheets[name] = s
	return nil
}
