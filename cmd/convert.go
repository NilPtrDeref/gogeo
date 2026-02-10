package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nilptrderef/gogeo/internal/shapefile"
	"github.com/nilptrderef/gogeo/internal/simplification"
	"github.com/tinylib/msgp/msgp"

	"github.com/spf13/cobra"
)

var (
	ShpPath            string
	DbfPath            string
	SimplifyPercentage float64
	SimplifyAlgorithm  string
	OutFile            string
)

var ConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a shapefile, and optionally a '.dbf' file into GeoJSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		var simplifier simplification.Simplifier
		if cmd.Flags().Changed("sp") {
			switch SimplifyAlgorithm {
			case "vis":
				simplifier = simplification.VisvalingamSimplifier{}
			case "doug":
				simplifier = simplification.DouglasPeuckerSimplifier{}
			default:
				return fmt.Errorf("invalid settings")
			}
		}

		file, err := os.Open(ShpPath)
		if err != nil {
			return err
		}
		defer file.Close()

		shp, err := shapefile.Parse(file)
		if err != nil {
			return err
		}

		if DbfPath != "" {
			dfile, err := os.Open(DbfPath)
			if err != nil {
				return err
			}
			defer dfile.Close()

			if err := shp.LoadAttributes(dfile); err != nil {
				return err
			}
		}

		var out *os.File
		if OutFile != "" {
			out, err = os.Create(OutFile)
			if err != nil {
				return err
			}
		} else {
			out = os.Stdout
		}

		if strings.HasSuffix(OutFile, "msgpk") {
			counties := shp.ToCounties()
			err = counties.SimplifyInPlace(simplifier, SimplifyPercentage)
			if err != nil {
				return err
			}

			writer := msgp.NewWriter(out)
			err = counties.EncodeMsg(writer)
			if err != nil {
				return err
			}
			return writer.Flush()
		}

		geojson := shp.ToGeoJson()
		err = geojson.SimplifyInPlace(simplifier, SimplifyPercentage)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(out)
		return encoder.Encode(geojson)
	},
}

func init() {
	ConvertCmd.Flags().StringVarP(&ShpPath, "shp", "s", "", "Path of the shapefile")
	ConvertCmd.MarkFlagRequired("shp")
	ConvertCmd.Flags().StringVarP(&DbfPath, "dbf", "d", "", "Path of the dbase file")
	ConvertCmd.Flags().Float64VarP(&SimplifyPercentage, "sp", "p", 1.0, "A float between 0 and 1 that represents the approximate percentage of remaining points")
	ConvertCmd.Flags().StringVarP(&SimplifyAlgorithm, "sa", "a", "doug", "The algorithm to use when simplifying. 'vis' for Visvalingam-Whyatt or 'doug' for Douglas-Peucker)")
	ConvertCmd.Flags().StringVarP(&OutFile, "output", "o", "", "Output file path")
}
