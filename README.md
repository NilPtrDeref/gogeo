# gogeo

Golang Geography and US Map visualization Application

## Web interface

You can serve a web interface to view the map. It provides some basic zoom/move functionality.
Currently, it expects that you have converted the shapefile to a msgpack file and pre-projected
the points.

It's not slow at 100% of the geometry, but it does load much faster when you simplify the geometry
down to about 10%.

## Data Sources

Download the County or ZCTA (ZIP Code Tabulation Areas) shapefiles from the
[US Census Bureau](https://www.census.gov/geographies/mapping-files/time-series/geo/tiger-line-file.html)
website for the year that you'd like. It will come in a zipfile containing
`.cpg`,`.dbf`,`.prj`,`.shp`,`.shx`, and a few `.xml` files.
The `.shp` and `.dbf` are the primary files that will be necessary for this
system. You'll process them using the `gogeo convert` command to prepare them
in order for the `gogeo serve` command to display them.
