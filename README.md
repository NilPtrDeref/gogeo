# gogeo

Golang Geography and US Map visualization Application

## Data Sources

Download the County or ZCTA (ZIP Code Tabulation Areas) shapefiles from the
[US Census Bureau](https://www.census.gov/geographies/mapping-files/time-series/geo/tiger-line-file.html)
website for the year that you'd like. It will come in a zipfile containing
`.cpg`,`.dbf`,`.prj`,`.shp`,`.shx`, and a few `.xml` files.
The `.shp` and `.dbf` are the primary files that will be necessary for this
system. You'll process them using the `gogeo convert` command to prepare them
in order for the `gogeo serve` command to display them.
