package gc70val

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/funwithbots/go-gedcom/internal"

	"gopkg.in/yaml.v3"
)

const (
	G7UCLetter   = "A-Z"
	G7Digit      = "0-9"
	G7Nonzero    = "1-9"
	G7Underscore = "_"
	G7Atsign     = "@"
	G7Banned     = `\x{0}-\x{8}\x{B}-\x{C}\x{E}-\x{1F}\x{7F}\x{80}-\x{9F}\x{D800}-\x{DFFF}\x{FFFE}-\x{FFFF}`

	abnfDir = "data/abnf"
)

var (
	baseline = struct {
		tags      map[string]TagDef
		calendars map[string]calDef
		types     map[string]typeDef
		enumSets  map[string]enumSet
	}{}
)

var (
// regLevel  = regexp.MustCompile(fmt.Sprintf("[0%s]+", g7Nonzero))
)

//go:embed data/abnf/*
var abnfFS embed.FS

func init() {
	var (
		tags      = pseudoTags
		types     = make(map[string]typeDef)
		calendars = make(map[string]calDef)
		enumSets  = make(map[string]enumSet)

		err error
	)

	// Environment variable to toggle debug logging.
	const (
		debugLogEnvKey = "GOGEDCOM_DEBUG"
		debugLogEnvVal = "1"
	)
	if os.Getenv(debugLogEnvKey) == debugLogEnvVal { // Reduce the footprint of this potentially global change.
		currentLogLevel := slog.SetLogLoggerLevel(slog.LevelDebug)
		defer slog.SetLogLoggerLevel(currentLogLevel)
	}
	logHandler := internal.NewEnvFilterHandler(
		slog.Default().Handler(),
		debugLogEnvKey, debugLogEnvVal,
	)
	// This logger will only output when the environment variable is set.
	logger := slog.New(logHandler).
		With(slog.String("logger", "GEDCOM Logger"))
	logger.Debug("Importing Gedcom 7 configs.")

	AddValidTag(TagHEAD)
	AddValidTag(TagTRLR)
	AddValidTag(TagCONT)
	logger.Debug("Added tags", slog.Any("tags", []string{TagHEAD, TagTRLR, TagCONT}))

	files, err := abnfFS.ReadDir(abnfDir)
	if err != nil {
		panic(fmt.Sprintf("unable to open abnf folder %s: %v", abnfDir, err))
	}

	for _, fn := range files {
		fileLogger := logger.With(slog.String("filename", fn.Name()))
		data, err := abnfFS.ReadFile(abnfDir + "/" + fn.Name())
		if err != nil {
			fileLogger.Debug("Reading file but got an error, skipping", slog.String("error", err.Error()))
			continue
		}
		fileLogger.Debug("Processing file")

		name := strings.Split(fn.Name(), "-")[0]
		switch name {
		case "enum":
			// These values are extracted from enumSets.
			continue
		case "month":
			// These values are extracted from calendars.
			continue
		case "enumset":
			if es, err := loadEnumSet(data); err != nil {
				fileLogger.Debug("Error parsing enumSet", slog.String("error", err.Error()))
			} else {
				enumSets[es.URI] = es
			}
		case "cal":
			cm, err := loadCal(data)
			if err != nil {
				fileLogger.Debug("Error parsing as calendar: %s", slog.String("error", err.Error()))
			} else {
				calendars[cm.Cal] = cm
				fileLogger.Debug("Added calendar", slog.String("cal", cm.Cal))
			}
		case "type":
			tm, err := loadType(data)
			if err != nil {
				fileLogger.Debug("Error parsing as type", slog.String("error", err.Error()))
			} else {
				types[tm.Type] = tm
				fileLogger.Debug("Added type", slog.String("type", tm.Type))
			}
		default:
			t, err := loadTag(data)
			if err != nil {
				fileLogger.Debug("Error parsing as default", slog.String("error", err.Error()))
			} else {
				if name == "ord" {
					// special case for LDS Ordinance tags
					t.FullTag = "ord-" + t.FullTag
				}
				if name == "record" {
					// record types have no superstructures and distinct substructure lists.
					// They apply only if level is 0.
					t.FullTag = "record-" + t.FullTag
				}
				tags[t.FullTag] = t
				fileLogger.Debug("Loaded tag", slog.String("tag", t.FullTag))
			}
		}
	}

	for key, tag := range tags {
		if tag.EnumSetName != "" {
			if es, ok := enumSets[tag.EnumSetName]; !ok {
				logger.Debug("No matching tag for %s to %s.\n", key, tag.EnumSetName)
			} else {
				tag.EnumSet = es
				tags[key] = tag
				logger.Debug("Added enumset", slog.String("tag", es.FullTag))
			}
		}
	}

	baseline.tags = tags
	baseline.calendars = calendars
	baseline.types = types
	baseline.enumSets = enumSets
}

// deserializeYAML populates v with the contents of the first document in the YAML text.
// It skips everything prior to the first document delimiter.
func deserializeYAML[T any](data []byte, v *T) error {
	pos := bytes.Index(data, []byte("---"))
	if pos == -1 {
		pos = 0
	}

	decoder := yaml.NewDecoder(bytes.NewBuffer(data[pos:]))
	for {
		if err := decoder.Decode(v); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("document decode failed: %w", err)
		}
	}
	return nil
}
