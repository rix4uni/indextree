## indextree

A Go CLI tool that recursively crawls HTTP directory listings (such as Apache autoindex pages) and prints the structure in a pretty tree format (similar to the standard `tree` command), enriched with file sizes, last modified times, and terminal colors.

## Features
- **Piped Stdin Input**: Consumes directories from `stdin`, allowing you to crawl multiple targets sequentially.
- **File Size & Metadata Parsing**: Extracts sizes (e.g. `8.7K`, `3.6M`) and last modified times from listing rows.
- **Deadlock-Free Concurrency**: Leverages goroutines and channel-based semaphores to fetch subfolders concurrently.
- **Resilient Request Retries**: Automatically retries on rate limits (HTTP `429`) or server errors (HTTP `5xx`) using exponential backoff.
- **Auto-Formatting**: Colors are automatically stripped when output is redirected to a file.

## Installation
```
go install github.com/rix4uni/indextree@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/indextree/releases/download/v0.0.1/indextree-linux-amd64-0.0.1.tgz
tar -xvzf indextree-linux-amd64-0.0.1.tgz
rm -rf indextree-linux-amd64-0.0.1.tgz
mv indextree ~/go/bin/indextree
```
Or download [binary release](https://github.com/rix4uni/indextree/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/indextree.git
cd indextree; go install
```

## Usage
```console
Usage of indextree:
      --color string      Color output (always, never, auto) (default "auto")
      --concurrency int   Maximum number of concurrent HTTP requests (default 50)
      --delay int         Delay between requests in milliseconds (default 100)
      --depth int         Maximum recursion depth (0 for unlimited)
      --output string     Output file path (default is stdout)
      --retries int       Number of retries for failed HTTP requests (default 3)
      --silent            Silent mode.
      --verbose           Verbose output (show crawl details/progress)
      --version           Print the version of the tool and exit.
```

## Examples

### 1. Simple Directory Listing (Depth 1)
```console
echo "https://spdf.gsfc.example.com/pub/" | indextree --depth 2
```

**Output**:
```console
pub/
├── catalogs/
│   ├── plot_walk/ (2025-09-19 14:49)
│   ├── 00readme.txt (3.6K, 2024-05-07 11:05)
│   ├── all.xml (6.7M, 2026-07-21 09:18)
│   ├── CDAWebdataset-dirs.txt (268K, 2026-07-21 16:29)
│   ├── cdfmetafile_for_hapi.txt (1.8G, 2026-07-21 09:19)
│   ├── filelist.gz (1.1G, 2026-07-17 23:42)
│   ├── filelist.patch.gz (1.0G, 2026-07-17 23:29)
│   ├── filelist.txt (794, 2023-02-13 21:37)
│   ├── spdf-plotwalk-catalog.json (138K, 2026-07-18 05:15)
│   ├── spdf-plotwalk-dates.json (105K, 2026-07-18 05:15)
│   └── spdf_filelist.sh (2.0K, 2023-11-03 22:25)
├── data/
│   ├── 1963-038C/ (2014-08-08 11:05)
│   ├── aaa_balloons/ (2021-01-13 15:36)
│   ├── aaa_groundbased/ (2023-04-05 16:15)
│   ├── aaa_planetary/ (2026-02-05 11:48)
│   ├── aaa_smallsats_cubesats/ (2026-02-24 14:35)
│   ├── aaa_sounding_rockets/ (2023-03-10 14:33)
│   ├── aaa_special-purpose-datasets/ (2024-05-07 13:18)
│   ├── ace/ (2025-02-10 00:48)
│   ├── ae/ (2013-03-04 14:58)
│   ├── aeros/ (2013-12-03 08:51)
│   ├── aim/ (2021-01-29 14:58)
│   ├── alouette/ (2015-06-19 10:41)
│   ├── ampte/ (2009-02-12 11:33)
│   ├── apollo/ (2021-09-24 21:21)
│   ├── arase/ (2019-05-07 13:35)
│   ├── arcad/ (2023-08-14 11:44)
│   ├── ariel/ (2014-05-28 15:36)
│   ├── ats/ (2012-09-26 14:20)
│   ├── aureol/ (2017-05-11 14:34)
│   ├── awe/ (2025-07-09 15:12)
│   ├── azur/ (2014-08-07 15:25)
│   ├── balloons/ (2021-01-13 15:36)
│   ├── barrel/ (2020-12-15 22:08)
│   ├── bepicolombo/ (2025-02-10 00:48)
│   ├── canopus/ (2012-10-01 18:08)
│   ├── cassini/ (2025-02-10 00:48)
│   ├── cdaw9/ (2012-10-01 01:03)
│   ├── cluster/ (2024-06-12 19:42)
│   ├── cnofs/ (2013-02-13 16:37)
│   ├── cosmos/ (2014-10-02 11:03)
│   ├── crres/ (2012-09-29 22:59)
│   ├── csswe/ (2014-02-21 00:16)
│   ├── darn/ (2012-10-01 21:25)
│   ├── dawn/ (2025-02-10 00:48)
│   ├── de/ (2025-04-01 10:12)
│   ├── dmsp/ (2019-01-10 18:12)
│   ├── doublestar/ (2015-12-30 13:55)
│   ├── dscovr/ (2025-02-10 00:48)
│   ├── dsp/ (2014-10-02 11:29)
│   ├── elfin/ (2023-10-10 18:49)
│   ├── equator-s/ (2012-09-29 21:22)
│   ├── erg/ (2019-05-07 13:35)
│   ├── ers/ (2021-09-24 21:38)
│   ├── esa/ (2014-09-24 10:24)
│   ├── esro/ (2014-09-23 07:27)
│   ├── explorer/ (2024-10-24 17:05)
│   ├── fast/ (2020-06-29 14:31)
│   ├── finnish_meteor_inst_fmi/ (2012-09-29 21:28)
│   ├── formosat-rocsat/ (2020-12-17 13:12)
│   ├── galileo/ (2026-01-15 15:28)
│   ├── genesis/ (2012-10-04 15:37)
│   ├── geotail/ (2021-12-23 15:05)
│   ├── giotto/ (2025-02-10 00:48)
│   ├── goes/ (2025-11-11 00:18)
│   ├── gold/ (2024-01-31 12:58)
│   ├── gps/ (2021-10-25 02:21)
│   ├── hawkeye/ (2026-04-28 16:20)
│   ├── heao/ (2017-01-26 15:27)
│   ├── helios/ (2012-09-26 11:57)
│   ├── heos/ (2007-11-29 13:38)
│   ├── hinotori/ (2001-02-12 11:31)
│   ├── ibex/ (2023-07-27 23:04)
│   ├── icon/ (2026-07-13 14:46)
│   ├── image/ (2023-08-14 11:44)
│   ├── imap/ (2026-02-20 09:49)
│   ├── imp/ (2014-10-15 18:44)
│   ├── injun/ (2014-08-08 15:09)
│   ├── interball/ (2021-12-23 15:23)
│   ├── international_space_station_iss/ (2024-09-26 16:07)
│   ├── interstellar_probe/ (2019-02-13 12:42)
│   ├── ionosphere_sounding_satellite_iss-2/ (2021-09-24 22:13)
│   ├── isee/ (2014-10-16 10:02)
│   ├── isis/ (2016-08-31 15:33)
│   ├── iue/ (2014-09-26 13:05)
│   ├── juice/ (2025-02-10 00:48)
│   ├── juno/ (2025-02-10 00:48)
│   ├── lanl/ (2021-12-23 15:31)
│   ├── lws-set/ (2021-11-23 17:19)
│   ├── magsat/ (2014-09-24 11:38)
│   ├── map/ (2012-10-01 00:39)
│   ├── mariner/ (2022-10-26 13:40)
│   ├── maven/ (2025-02-10 00:48)
│   ├── messenger/ (2025-02-10 00:48)
│   ├── meteosat/ (2017-05-15 14:57)
│   ├── mms/ (2023-07-18 15:55)
│   ├── munin/ (2017-08-28 13:13)
│   ├── new-horizons/ (2025-02-10 00:48)
│   ├── noaa/ (2019-08-06 16:48)
│   ├── ogo/ (2010-08-04 10:13)
│   ├── ohzora/ (2017-03-17 15:44)
│   ├── omni/ (2023-08-14 11:44)
│   ├── oso/ (2012-04-22 20:42)
│   ├── ov/ (2014-10-16 11:03)
│   ├── phobos/ (2017-08-09 06:48)
│   ├── pioneer/ (2025-01-29 15:10)
│   ├── planet_and_comet_positions/ (2025-02-10 00:48)
│   ├── pointer_files/ (2023-03-02 11:09)
│   ├── polar/ (2026-03-23 15:59)
│   ├── prognoz/ (2012-10-01 19:33)
│   ├── psp/ (2021-11-05 10:33)
│   ├── psyche/ (2025-02-10 00:48)
│   ├── radsat/ (2014-09-26 16:19)
│   ├── raids/ (2021-09-24 22:51)
│   ├── rbsp/ (2021-06-25 21:47)
│   ├── reach/ (2023-07-24 03:28)
│   ├── real/ (2026-06-25 16:56)
│   ├── relay/ (2008-06-24 12:11)
│   ├── rosetta/ (2025-02-10 00:48)
│   ├── russian_msu/ (2023-08-14 11:44)
│   ├── s15/ (2007-08-28 17:11)
│   ├── s3/ (2017-05-12 15:52)
│   ├── s_cubed_a/ (2024-09-10 13:29)
│   ├── sakigake/ (2025-02-10 00:48)
│   ├── sampex/ (2023-08-14 11:44)
│   ├── san_marco/ (2023-07-27 15:32)
│   ├── sesame/ (2012-10-01 18:11)
│   ├── shimmer/ (2023-08-14 11:44)
│   ├── skylab/ (2017-05-01 14:41)
│   ├── sme/ (2014-09-04 08:39)
│   ├── snoe/ (2012-09-29 21:32)
│   ├── soho/ (2025-02-10 00:48)
│   ├── solar/ (2026-02-24 15:17)
│   ├── solar-orbiter/ (2025-06-26 00:11)
│   ├── solar_maximum_mission/ (2014-09-05 13:10)
│   ├── solrad/ (2014-08-04 11:46)
│   ├── sondestrom/ (2012-10-01 01:03)
│   ├── sounding_rockets/ (2023-03-10 14:33)
│   ├── spartan/ (2017-05-02 12:12)
│   ├── st5/ (2014-12-15 02:36)
│   ├── stelab/ (2012-09-29 08:05)
│   ├── stereo/ (2022-08-23 21:23)
│   ├── stp/ (2026-04-02 15:57)
│   ├── suisei/ (2025-02-10 00:48)
│   ├── telstar/ (2014-08-07 09:12)
│   ├── themis/ (2019-03-27 12:37)
│   ├── timed/ (2024-05-07 16:37)
│   ├── tip/ (2014-09-26 16:57)
│   ├── tracers/ (2026-04-07 18:05)
│   ├── transit-5e5_1964-083c/ (2007-08-28 17:11)
│   ├── tss/ (2018-05-01 18:08)
│   ├── twins/ (2021-08-02 12:25)
│   ├── ulysses/ (2025-02-10 00:48)
│   ├── vanguard/ (2009-04-22 12:00)
│   ├── vega/ (2021-09-29 08:44)
│   ├── vela/ (2014-09-15 14:40)
│   ├── viking/ (2017-05-12 15:42)
│   ├── voyager/ (2021-01-29 15:18)
│   ├── wind/ (2025-02-10 00:48)
│   ├── zzz_old/ (2014-10-16 13:30)
│   ├── zzz_testing/ (2012-10-12 08:18)
│   └── 000_readme.txt (3.7K, 2023-08-11 14:47)
├── documents/
│   ├── archives-history/ (2022-10-07 16:58)
│   ├── HPDE/ (2018-06-04 12:59)
│   ├── metadata/ (2025-09-24 16:07)
│   ├── mission_PDMP_CMAD/ (2021-09-28 17:11)
│   ├── old/ (2024-11-08 14:55)
│   ├── research/ (2023-11-22 13:38)
│   ├── space_weathering/ (2009-09-14 14:46)
│   ├── SPDF/ (2026-07-08 16:15)
│   ├── spds/ (2006-06-29 15:33)
│   └── attribute_files.txt (5.6K, 2000-10-23 12:09)
├── event_lists/
│   ├── carrington_event_1859_auroral_sightings/ (2023-07-31 09:44)
│   ├── other-event-lists/ (2024-05-07 13:18)
│   ├── 00readme.txt (1.1K, 2022-04-11 18:24)
│   ├── Berdichevsky - changes in the direction of the Ey component of the electric field (26K, 2025-07-11 12:30)
│   ├── Berdichevsky - interactive regions in the SW (10K, 2025-07-11 12:30)
│   ├── Berdichevsky - Interplanetary magnetic clouds (8.0K, 2025-07-11 12:30)
│   ├── Berdichevsky - interplanetary shocks (20K, 2025-07-11 12:30)
│   ├── Berdichevsky - intervals of IMF Bz northward (22K, 2025-07-11 12:30)
│   ├── Berdichevsky - intervals of IMF Bz southward (31K, 2025-07-11 12:30)
│   ├── Berdichevsky - low speed stream intervals in the SW (6.2K, 2025-07-11 12:30)
│   ├── Berdichevsky - magnetic field sector boundary crossings (21K, 2025-07-11 12:30)
│   ├── Berdichevsky - miscelaneous features in the SW (25K, 2025-07-11 12:30)
│   ├── Berdichevsky - SW high speed streams (8.6K, 2025-07-11 12:30)
│   ├── eriksson_wind_current_sheet_exhaust_lmn_v01.txt (437K, 2022-04-11 18:25)
│   └── Lepping and Berdichevsky - ISTP Solar Wind Catalog.txt (149K, 2025-07-11 12:43)
├── misc/
│   ├── bob_isis/ (2016-07-07 16:36)
│   ├── cdaw6_9/ (2014-04-07 09:30)
│   ├── old/ (2014-10-16 13:21)
│   ├── orbits/ (2012-10-03 14:32)
│   └── photo_gallery/ (2009-04-14 14:11)
├── models/
│   ├── old_nonpublic/ (2017-11-13 15:47)
│   └── orbital_radiation_package_orp/ (2022-02-18 14:09)
├── pre_generated_plots/
│   ├── ace/ (2013-03-12 13:59)
│   ├── cnofs/ (2012-04-05 09:04)
│   ├── de/ (2021-05-24 22:10)
│   ├── geotail/ (2012-09-06 11:56)
│   ├── image/ (2013-08-27 12:06)
│   ├── isee/ (2014-12-10 16:39)
│   ├── kp_plots/ (2017-01-04 07:55)
│   ├── mms/ (2018-03-19 13:34)
│   ├── orbits/ (2025-02-24 16:03)
│   ├── other/ (2008-10-29 15:29)
│   ├── polar/ (2012-05-31 11:30)
│   ├── rbsp/ (2022-11-29 12:22)
│   ├── twins/ (2013-07-03 09:36)
│   └── ulysses/ (2013-04-11 06:43)
├── software/
│   ├── cdaweb_idl_clients/ (2021-11-04 11:26)
│   ├── cdawlib/ (2022-04-08 16:43)
│   ├── cdf/ (2024-10-01 08:26)
│   ├── format_conversion/ (2025-01-16 12:30)
│   └── old/ (2023-03-27 13:59)
├── 000_readme.htm (8.7K, 2024-10-09 16:16)
├── 000_readme.txt (5.6K, 2024-10-09 16:18)
├── ALL_FILES_ARE_PUBLIC (130, 2024-10-09 14:56)
└── datasets.html (3.6K, 2022-09-07 16:22)
```

### 2. Crawl Subdirectories (Depth 2) with Concurrency and Log Verbosity
```console
echo "https://spdf.gsfc.example.com/pub/software" | indextree --depth 2 --verbose
```

### 3. Save Tree to a Text File (Colors Automatically Disabled)
```console
echo "https://spdf.gsfc.example.com/pub/software/" | indextree --depth 2 --output tree.txt
```