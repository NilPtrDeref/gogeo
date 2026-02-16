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

## Other Resources

* [US Census Geography Relationship Files](https://www.census.gov/geographies/reference-files/time-series/geo/relationship-files.2020.html)
* [D3 Geography Documentation](https://d3js.org/d3-geo)
* [D3 Conic Projection Documentation](https://d3js.org/d3-geo/conic)
* [Shapefile Specification Information](https://en.wikipedia.org/wiki/Shapefile)
* [DBase file Specification Information](https://www.clicketyclick.dk/databases/xbase/format/dbf.html)
* [Visvalingam-Whyatt Algorithm](https://en.wikipedia.org/wiki/Visvalingam%E2%80%93Whyatt_algorithm)
* [Douglas-Peucker Algorithm](https://en.wikipedia.org/wiki/Ramer%E2%80%93Douglas%E2%80%93Peucker_algorithm)
* [Housing and Urban Development Zipcode Data](https://www.huduser.gov/portal/datasets/usps_crosswalk.html)
