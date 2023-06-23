//nolint:funlen
package internal_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/slavsan/gocov/internal"
)

const exampleCoverageOut3 = `mode: atomic
github.com/slavsan/gocov/cmd/gocov.go:9.13,16.22 5 0
github.com/slavsan/gocov/cmd/gocov.go:29.2,37.3 1 0
github.com/slavsan/gocov/cmd/gocov.go:16.22,17.21 1 0
github.com/slavsan/gocov/cmd/gocov.go:18.16,19.28 1 0
github.com/slavsan/gocov/cmd/gocov.go:20.18,22.24 2 0
github.com/slavsan/gocov/cmd/gocov.go:22.24,24.5 1 0
`

const exampleCoverageOut2 = `mode: atomic
github.com/slavsan/gocov/main.go:5.13,7.2 1 0
github.com/slavsan/gocov/internal/gocov.go:44.52,58.15 4 2
github.com/slavsan/gocov/internal/gocov.go:60.2,63.64 4 2
github.com/slavsan/gocov/internal/gocov.go:66.2,66.21 1 2
github.com/slavsan/gocov/internal/gocov.go:89.2,89.26 1 2
github.com/slavsan/gocov/internal/gocov.go:95.2,95.26 1 2
github.com/slavsan/gocov/internal/gocov.go:58.15,58.32 1 2
github.com/slavsan/gocov/internal/gocov.go:63.64,64.33 1 0
github.com/slavsan/gocov/internal/gocov.go:66.21,72.32 5 292
github.com/slavsan/gocov/internal/gocov.go:76.3,77.17 2 292
github.com/slavsan/gocov/internal/gocov.go:80.3,83.23 3 292
github.com/slavsan/gocov/internal/gocov.go:86.3,86.96 1 292
github.com/slavsan/gocov/internal/gocov.go:72.32,74.4 1 7
github.com/slavsan/gocov/internal/gocov.go:77.17,79.4 1 0
github.com/slavsan/gocov/internal/gocov.go:83.23,85.4 1 227
github.com/slavsan/gocov/internal/gocov.go:89.26,93.3 3 7
github.com/slavsan/gocov/internal/gocov.go:103.33,108.2 1 2
github.com/slavsan/gocov/internal/gocov.go:110.68,117.33 6 2
github.com/slavsan/gocov/internal/gocov.go:120.2,122.30 2 2
github.com/slavsan/gocov/internal/gocov.go:126.2,126.131 1 2
github.com/slavsan/gocov/internal/gocov.go:117.33,119.3 1 2
github.com/slavsan/gocov/internal/gocov.go:122.30,125.3 2 2
github.com/slavsan/gocov/internal/gocov.go:129.40,134.2 4 2
github.com/slavsan/gocov/internal/gocov.go:136.60,138.20 2 14
github.com/slavsan/gocov/internal/gocov.go:142.2,142.32 1 14
github.com/slavsan/gocov/internal/gocov.go:152.2,155.32 4 14
github.com/slavsan/gocov/internal/gocov.go:158.2,159.38 2 14
github.com/slavsan/gocov/internal/gocov.go:162.2,162.52 1 14
github.com/slavsan/gocov/internal/gocov.go:138.20,141.3 2 7
github.com/slavsan/gocov/internal/gocov.go:142.32,145.33 3 12
github.com/slavsan/gocov/internal/gocov.go:148.3,148.35 1 12
github.com/slavsan/gocov/internal/gocov.go:145.33,147.4 1 9
github.com/slavsan/gocov/internal/gocov.go:148.35,150.4 1 9
github.com/slavsan/gocov/internal/gocov.go:155.32,157.3 1 7
github.com/slavsan/gocov/internal/gocov.go:159.38,161.3 1 7
github.com/slavsan/gocov/internal/gocov.go:165.75,167.2 1 12
github.com/slavsan/gocov/internal/gocov.go:169.97,173.19 4 12
github.com/slavsan/gocov/internal/gocov.go:178.2,178.19 1 12
github.com/slavsan/gocov/internal/gocov.go:182.2,190.28 4 12
github.com/slavsan/gocov/internal/gocov.go:193.2,194.30 2 12
github.com/slavsan/gocov/internal/gocov.go:173.19,175.3 1 5
github.com/slavsan/gocov/internal/gocov.go:175.8,175.26 1 7
github.com/slavsan/gocov/internal/gocov.go:175.26,177.3 1 2
github.com/slavsan/gocov/internal/gocov.go:178.19,181.3 2 12
github.com/slavsan/gocov/internal/gocov.go:190.28,192.3 1 10
github.com/slavsan/gocov/internal/gocov.go:194.30,197.3 2 10
github.com/slavsan/gocov/internal/gocov.go:200.34,202.2 1 12
github.com/slavsan/gocov/internal/gocov.go:204.49,206.2 1 7
github.com/slavsan/gocov/internal/gocov.go:216.49,219.15 2 17
github.com/slavsan/gocov/internal/gocov.go:226.2,226.44 1 10
github.com/slavsan/gocov/internal/gocov.go:229.2,229.53 1 10
github.com/slavsan/gocov/internal/gocov.go:219.15,220.37 1 7
github.com/slavsan/gocov/internal/gocov.go:223.3,223.9 1 7
github.com/slavsan/gocov/internal/gocov.go:220.37,222.4 1 7
github.com/slavsan/gocov/internal/gocov.go:226.44,228.3 1 5
github.com/slavsan/gocov/internal/gocov.go:232.65,234.22 2 2
github.com/slavsan/gocov/internal/gocov.go:237.2,238.46 2 2
github.com/slavsan/gocov/internal/gocov.go:234.22,236.3 1 7
github.com/slavsan/gocov/internal/gocov.go:241.35,244.15 3 2
github.com/slavsan/gocov/internal/gocov.go:245.2,248.41 4 2
github.com/slavsan/gocov/internal/gocov.go:251.2,251.44 1 2
github.com/slavsan/gocov/internal/gocov.go:244.15,244.32 1 2
github.com/slavsan/gocov/internal/gocov.go:248.41,249.31 1 0
github.com/slavsan/gocov/internal/gocov.go:254.53,260.51 2 292
github.com/slavsan/gocov/internal/gocov.go:287.2,287.21 1 292
github.com/slavsan/gocov/internal/gocov.go:260.51,261.52 1 5028
github.com/slavsan/gocov/internal/gocov.go:264.3,265.17 2 1752
github.com/slavsan/gocov/internal/gocov.go:268.3,270.17 2 1752
github.com/slavsan/gocov/internal/gocov.go:285.3,285.11 1 1752
github.com/slavsan/gocov/internal/gocov.go:261.52,262.12 1 3276
github.com/slavsan/gocov/internal/gocov.go:265.17,267.4 1 0
github.com/slavsan/gocov/internal/gocov.go:271.10,272.29 1 292
github.com/slavsan/gocov/internal/gocov.go:273.10,274.31 1 292
github.com/slavsan/gocov/internal/gocov.go:275.10,276.27 1 292
github.com/slavsan/gocov/internal/gocov.go:277.10,278.29 1 292
github.com/slavsan/gocov/internal/gocov.go:279.10,280.35 1 292
github.com/slavsan/gocov/internal/gocov.go:281.10,282.24 1 292
github.com/slavsan/gocov/internal/gocov.go:290.23,291.16 1 4
github.com/slavsan/gocov/internal/gocov.go:291.16,292.13 1 0
github.com/slavsan/gocov/internal/gocov.go:296.31,297.14 1 52
github.com/slavsan/gocov/internal/gocov.go:300.2,301.15 2 42
github.com/slavsan/gocov/internal/gocov.go:305.2,305.15 1 42
github.com/slavsan/gocov/internal/gocov.go:297.14,299.3 1 10
github.com/slavsan/gocov/internal/gocov.go:301.15,304.3 2 104
github.com/slavsan/gocov/cmd/gocov.go:9.13,13.2 3 0
`

const exampleCoverageOut = `mode: atomic
github.com/slavsan/gospec/cmd/cover.go:11.13,20.15 5 0
github.com/slavsan/gospec/cmd/cover.go:22.2,27.21 4 0
github.com/slavsan/gospec/cmd/cover.go:50.2,52.97 3 0
github.com/slavsan/gospec/cmd/cover.go:20.15,20.32 1 0
github.com/slavsan/gospec/cmd/cover.go:27.21,33.47 3 0
github.com/slavsan/gospec/cmd/cover.go:33.47,44.16 9 0
github.com/slavsan/gospec/cmd/cover.go:44.16,46.5 1 0
github.com/slavsan/gospec/cmd/cover.go:55.23,56.16 1 0
github.com/slavsan/gospec/cmd/cover.go:56.16,57.13 1 0
github.com/slavsan/gospec/expect.go:26.46,32.18 2 122
github.com/slavsan/gospec/expect.go:32.19,34.7 0 0
github.com/slavsan/gospec/expect.go:40.34,42.6 0 0
github.com/slavsan/gospec/expect.go:43.30,45.6 0 0
github.com/slavsan/gospec/expect.go:47.38,50.5 0 0
github.com/slavsan/gospec/expect.go:52.32,57.105 3 66
github.com/slavsan/gospec/expect.go:62.6,62.32 1 65
github.com/slavsan/gospec/expect.go:70.6,71.38 2 61
github.com/slavsan/gospec/expect.go:57.105,60.7 2 1
github.com/slavsan/gospec/expect.go:62.32,64.39 2 4
github.com/slavsan/gospec/expect.go:67.7,67.13 1 4
github.com/slavsan/gospec/expect.go:64.39,66.8 1 1
github.com/slavsan/gospec/expect.go:71.38,73.7 1 1
github.com/slavsan/gospec/expect.go:75.31,77.6 0 0
github.com/slavsan/gospec/expect.go:81.32,83.7 0 0
github.com/slavsan/gospec/expect.go:85.17,87.28 2 7
github.com/slavsan/gospec/expect.go:90.6,90.28 1 5
github.com/slavsan/gospec/expect.go:94.6,94.66 1 1
github.com/slavsan/gospec/expect.go:87.28,89.7 1 2
github.com/slavsan/gospec/expect.go:91.84,92.13 1 4
github.com/slavsan/gospec/expect.go:96.18,99.13 3 4
github.com/slavsan/gospec/expect.go:103.6,103.20 1 3
github.com/slavsan/gospec/expect.go:99.13,102.7 2 1
github.com/slavsan/gospec/expect.go:103.20,105.7 1 1
github.com/slavsan/gospec/expect.go:107.19,110.13 3 3
github.com/slavsan/gospec/expect.go:114.6,114.20 1 2
github.com/slavsan/gospec/expect.go:110.13,113.7 2 1
github.com/slavsan/gospec/expect.go:114.20,116.7 1 1
github.com/slavsan/gospec/expect.go:118.33,120.45 2 42
github.com/slavsan/gospec/expect.go:120.45,123.37 3 6
github.com/slavsan/gospec/expect.go:127.7,127.97 1 4
github.com/slavsan/gospec/expect.go:123.37,126.8 2 2
github.com/slavsan/gospec/expect.go:135.53,141.18 2 0
github.com/slavsan/gospec/expect.go:141.19,143.7 0 0
github.com/slavsan/gospec/expect.go:149.34,151.6 0 0
github.com/slavsan/gospec/expect.go:152.30,154.6 0 0
github.com/slavsan/gospec/expect.go:156.38,159.5 0 0
github.com/slavsan/gospec/expect.go:161.32,166.105 3 0
github.com/slavsan/gospec/expect.go:171.6,171.32 1 0
github.com/slavsan/gospec/expect.go:179.6,180.38 2 0
github.com/slavsan/gospec/expect.go:166.105,169.7 2 0
github.com/slavsan/gospec/expect.go:171.32,173.39 2 0
github.com/slavsan/gospec/expect.go:176.7,176.13 1 0
github.com/slavsan/gospec/expect.go:173.39,175.8 1 0
github.com/slavsan/gospec/expect.go:180.38,182.7 1 0
github.com/slavsan/gospec/expect.go:184.31,186.6 0 0
github.com/slavsan/gospec/expect.go:190.32,192.7 0 0
github.com/slavsan/gospec/expect.go:194.17,196.28 2 0
github.com/slavsan/gospec/expect.go:199.6,199.28 1 0
github.com/slavsan/gospec/expect.go:203.6,203.66 1 0
github.com/slavsan/gospec/expect.go:196.28,198.7 1 0
github.com/slavsan/gospec/expect.go:200.84,201.13 1 0
github.com/slavsan/gospec/expect.go:205.18,208.13 3 0
github.com/slavsan/gospec/expect.go:212.6,212.20 1 0
github.com/slavsan/gospec/expect.go:208.13,211.7 2 0
github.com/slavsan/gospec/expect.go:212.20,214.7 1 0
github.com/slavsan/gospec/expect.go:216.19,219.13 3 0
github.com/slavsan/gospec/expect.go:223.6,223.20 1 0
github.com/slavsan/gospec/expect.go:219.13,222.7 2 0
github.com/slavsan/gospec/expect.go:223.20,225.7 1 0
github.com/slavsan/gospec/expect.go:227.33,229.45 2 0
github.com/slavsan/gospec/expect.go:229.45,232.37 3 0
github.com/slavsan/gospec/expect.go:236.7,236.97 1 0
github.com/slavsan/gospec/expect.go:232.37,235.8 2 0
github.com/slavsan/gospec/featurespec.go:41.50,48.2 3 1
github.com/slavsan/gospec/featurespec.go:50.57,58.2 2 1
github.com/slavsan/gospec/featurespec.go:60.60,68.2 2 1
github.com/slavsan/gospec/featurespec.go:70.58,78.2 2 2
github.com/slavsan/gospec/featurespec.go:80.55,88.2 2 5
github.com/slavsan/gospec/featurespec.go:90.54,96.2 1 2
github.com/slavsan/gospec/featurespec.go:98.54,104.2 1 2
github.com/slavsan/gospec/featurespec.go:106.67,111.14 1 1
github.com/slavsan/gospec/featurespec.go:111.14,119.38 2 1
github.com/slavsan/gospec/featurespec.go:124.4,127.30 3 1
github.com/slavsan/gospec/featurespec.go:131.4,133.38 2 1
github.com/slavsan/gospec/featurespec.go:175.4,177.30 3 1
github.com/slavsan/gospec/featurespec.go:183.4,185.27 2 1
github.com/slavsan/gospec/featurespec.go:197.4,197.33 1 1
github.com/slavsan/gospec/featurespec.go:119.38,120.33 1 0
github.com/slavsan/gospec/featurespec.go:121.5,121.11 1 0
github.com/slavsan/gospec/featurespec.go:127.30,129.5 1 2
github.com/slavsan/gospec/featurespec.go:133.38,135.38 2 2
github.com/slavsan/gospec/featurespec.go:135.38,138.40 3 2
github.com/slavsan/gospec/featurespec.go:172.6,172.30 1 2
github.com/slavsan/gospec/featurespec.go:138.40,142.14 4 6
github.com/slavsan/gospec/featurespec.go:145.7,145.32 1 4
github.com/slavsan/gospec/featurespec.go:142.14,143.16 1 2
github.com/slavsan/gospec/featurespec.go:146.19,147.24 1 2
github.com/slavsan/gospec/featurespec.go:150.8,150.21 1 2
github.com/slavsan/gospec/featurespec.go:151.29,153.25 2 2
github.com/slavsan/gospec/featurespec.go:156.8,156.22 1 2
github.com/slavsan/gospec/featurespec.go:160.43,162.25 2 0
github.com/slavsan/gospec/featurespec.go:165.8,165.22 1 0
github.com/slavsan/gospec/featurespec.go:147.24,149.9 1 1
github.com/slavsan/gospec/featurespec.go:153.25,155.9 1 0
github.com/slavsan/gospec/featurespec.go:162.25,164.9 1 0
github.com/slavsan/gospec/featurespec.go:177.30,182.5 3 2
github.com/slavsan/gospec/featurespec.go:185.27,188.31 3 2
github.com/slavsan/gospec/featurespec.go:194.5,194.25 1 2
github.com/slavsan/gospec/featurespec.go:188.31,193.6 3 4
github.com/slavsan/gospec/featurespec.go:205.39,207.2 1 0
github.com/slavsan/gospec/featurespec.go:209.32,213.2 1 1
github.com/slavsan/gospec/featurespec.go:215.36,217.28 2 0
github.com/slavsan/gospec/featurespec.go:217.28,219.3 1 0
github.com/slavsan/gospec/featurespec.go:222.36,231.28 2 1
github.com/slavsan/gospec/featurespec.go:301.2,301.15 1 1
github.com/slavsan/gospec/featurespec.go:231.28,232.26 1 14
github.com/slavsan/gospec/featurespec.go:238.3,238.29 1 13
github.com/slavsan/gospec/featurespec.go:245.3,245.27 1 12
github.com/slavsan/gospec/featurespec.go:253.3,253.24 1 10
github.com/slavsan/gospec/featurespec.go:264.3,264.23 1 5
github.com/slavsan/gospec/featurespec.go:275.3,275.23 1 3
github.com/slavsan/gospec/featurespec.go:286.3,286.24 1 1
github.com/slavsan/gospec/featurespec.go:232.26,235.12 3 1
github.com/slavsan/gospec/featurespec.go:238.29,242.12 3 1
github.com/slavsan/gospec/featurespec.go:245.27,250.12 3 2
github.com/slavsan/gospec/featurespec.go:253.24,254.26 1 5
github.com/slavsan/gospec/featurespec.go:258.4,260.12 3 3
github.com/slavsan/gospec/featurespec.go:254.26,256.13 2 2
github.com/slavsan/gospec/featurespec.go:264.23,265.26 1 2
github.com/slavsan/gospec/featurespec.go:269.4,271.12 3 2
github.com/slavsan/gospec/featurespec.go:265.26,267.13 2 0
github.com/slavsan/gospec/featurespec.go:275.23,276.26 1 2
github.com/slavsan/gospec/featurespec.go:280.4,282.12 3 2
github.com/slavsan/gospec/featurespec.go:276.26,278.13 2 0
github.com/slavsan/gospec/featurespec.go:286.24,295.12 2 1
github.com/slavsan/gospec/gospec.go:36.40,42.2 1 27
github.com/slavsan/gospec/gospec.go:44.30,45.32 1 4
github.com/slavsan/gospec/gospec.go:55.2,55.18 1 4
github.com/slavsan/gospec/gospec.go:45.32,46.30 1 70
github.com/slavsan/gospec/gospec.go:49.3,49.22 1 55
github.com/slavsan/gospec/gospec.go:53.3,53.64 1 25
github.com/slavsan/gospec/gospec.go:46.30,47.12 1 15
github.com/slavsan/gospec/gospec.go:49.22,51.12 2 30
github.com/slavsan/gospec/gospec.go:58.45,63.22 4 26
github.com/slavsan/gospec/gospec.go:77.2,77.37 1 26
github.com/slavsan/gospec/gospec.go:82.2,82.48 1 26
github.com/slavsan/gospec/gospec.go:90.2,90.44 1 26
github.com/slavsan/gospec/gospec.go:102.2,102.49 1 26
github.com/slavsan/gospec/gospec.go:112.2,112.40 1 26
github.com/slavsan/gospec/gospec.go:162.2,162.20 1 26
github.com/slavsan/gospec/gospec.go:169.2,169.15 1 26
github.com/slavsan/gospec/gospec.go:63.22,65.22 2 91
github.com/slavsan/gospec/gospec.go:71.3,71.30 1 86
github.com/slavsan/gospec/gospec.go:74.3,74.33 1 86
github.com/slavsan/gospec/gospec.go:65.22,67.58 2 65
github.com/slavsan/gospec/gospec.go:67.58,69.5 1 5
github.com/slavsan/gospec/gospec.go:71.30,73.4 1 299
github.com/slavsan/gospec/gospec.go:77.37,80.3 2 26
github.com/slavsan/gospec/gospec.go:82.48,83.23 1 11
github.com/slavsan/gospec/gospec.go:86.3,87.23 2 9
github.com/slavsan/gospec/gospec.go:83.23,85.4 1 2
github.com/slavsan/gospec/gospec.go:90.44,92.40 2 45
github.com/slavsan/gospec/gospec.go:99.3,99.27 1 45
github.com/slavsan/gospec/gospec.go:92.40,95.59 3 87
github.com/slavsan/gospec/gospec.go:95.59,96.10 1 45
github.com/slavsan/gospec/gospec.go:102.49,103.29 1 61
github.com/slavsan/gospec/gospec.go:109.3,109.15 1 56
github.com/slavsan/gospec/gospec.go:103.29,105.59 2 53
github.com/slavsan/gospec/gospec.go:105.59,107.5 1 5
github.com/slavsan/gospec/gospec.go:112.40,114.49 2 182
github.com/slavsan/gospec/gospec.go:126.3,126.50 1 179
github.com/slavsan/gospec/gospec.go:134.3,134.49 1 156
github.com/slavsan/gospec/gospec.go:142.3,142.30 1 95
github.com/slavsan/gospec/gospec.go:146.3,146.22 1 68
github.com/slavsan/gospec/gospec.go:114.49,115.30 1 3
github.com/slavsan/gospec/gospec.go:118.4,119.30 2 3
github.com/slavsan/gospec/gospec.go:123.4,124.12 2 3
github.com/slavsan/gospec/gospec.go:115.30,117.5 1 3
github.com/slavsan/gospec/gospec.go:119.30,122.5 2 3
github.com/slavsan/gospec/gospec.go:126.50,127.30 1 23
github.com/slavsan/gospec/gospec.go:130.4,132.12 3 23
github.com/slavsan/gospec/gospec.go:127.30,129.5 1 5
github.com/slavsan/gospec/gospec.go:134.49,137.31 3 61
github.com/slavsan/gospec/gospec.go:140.4,140.12 1 61
github.com/slavsan/gospec/gospec.go:137.31,139.5 1 5
github.com/slavsan/gospec/gospec.go:142.30,144.12 2 27
github.com/slavsan/gospec/gospec.go:146.22,150.30 4 68
github.com/slavsan/gospec/gospec.go:158.4,158.12 1 68
github.com/slavsan/gospec/gospec.go:150.30,152.67 2 52
github.com/slavsan/gospec/gospec.go:152.67,154.6 1 3
github.com/slavsan/gospec/gospec.go:155.10,157.5 1 16
github.com/slavsan/gospec/gospec.go:162.20,163.59 1 17
github.com/slavsan/gospec/gospec.go:163.60,164.4 0 7
github.com/slavsan/gospec/gospec.go:164.9,166.4 1 10
github.com/slavsan/gospec/gospec.go:172.29,179.36 3 4
github.com/slavsan/gospec/gospec.go:179.36,180.63 1 32
github.com/slavsan/gospec/gospec.go:180.63,181.36 1 32
github.com/slavsan/gospec/gospec.go:181.36,183.5 1 133
github.com/slavsan/gospec/gospec.go:188.44,190.26 2 59
github.com/slavsan/gospec/gospec.go:198.2,198.20 1 59
github.com/slavsan/gospec/gospec.go:190.26,191.47 1 216
github.com/slavsan/gospec/gospec.go:191.47,192.14 1 172
github.com/slavsan/gospec/gospec.go:195.4,195.46 1 172
github.com/slavsan/gospec/gospec.go:192.14,194.5 1 113
github.com/slavsan/gospec/gospec.go:211.55,222.2 4 326
github.com/slavsan/gospec/gospec.go:224.43,230.2 1 225
github.com/slavsan/gospec/gospec.go:232.49,239.2 1 426
`

func TestStdoutReport(t *testing.T) {
	testCases := []struct {
		title            string
		fsys             fs.StatFS
		config           *internal.Config
		expectedStdout   string
		expectedStderr   string
		expectedExitCode int
	}{
		{
			title: "with example coverage.out file and stdout report and colors disabled",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|--------------------|---------|----------|`,
				`| File               |   Stmts |  % Stmts |`,
				`|--------------------|---------|----------|`,
				`| gospec             | 237/323 |   73.37% |`,
				`|   cmd              |    0/28 |    0.00% |`,
				`|     cover.go       |    0/28 |    0.00% |`,
				`|   expect.go        |   43/86 |   50.00% |`,
				`|   featurespec.go   |  92/107 |   85.98% |`,
				`|   gospec.go        | 102/102 |  100.00% |`,
				`|--------------------|---------|----------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with example coverage.out file and stdout report and colors enabled",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: true,
			},
			expectedStdout: strings.Join([]string{
				"|--------------------|---------|----------|",
				"| File               |   Stmts |  % Stmts |",
				"|--------------------|---------|----------|",
				"|\033[0;33m gospec             \033[0m| \033[0;33m237/323\033[0m | \033[0;33m  73.37%\033[0m |",
				"|\033[0;31m   cmd              \033[0m| \033[0;31m   0/28\033[0m | \033[0;31m   0.00%\033[0m |",
				"|\033[0;31m     cover.go       \033[0m| \033[0;31m   0/28\033[0m | \033[0;31m   0.00%\033[0m |",
				"|\033[0;33m   expect.go        \033[0m| \033[0;33m  43/86\033[0m | \033[0;33m  50.00%\033[0m |",
				"|\033[0;32m   featurespec.go   \033[0m| \033[0;32m 92/107\033[0m | \033[0;32m  85.98%\033[0m |",
				"|\033[0;32m   gospec.go        \033[0m| \033[0;32m102/102\033[0m | \033[0;32m 100.00%\033[0m |",
				"|--------------------|---------|----------|",
				"",
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with another example coverage.out file and stdout report",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut2)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|----------------|---------|----------|`,
				`| File           |   Stmts |  % Stmts |`,
				`|----------------|---------|----------|`,
				`| gocov          | 133/142 |   93.66% |`,
				`|   cmd          |     0/3 |    0.00% |`,
				`|     gocov.go   |     0/3 |    0.00% |`,
				`|   internal     | 133/138 |   96.38% |`,
				`|     gocov.go   | 133/138 |   96.38% |`,
				`|   main.go      |     0/1 |    0.00% |`,
				`|----------------|---------|----------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with .gocov file specifying one file to ignore",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut2)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"ignore": [`,
					`		"gocov/main.go"`,
					`	]`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|----------------|---------|----------|`,
				`| File           |   Stmts |  % Stmts |`,
				`|----------------|---------|----------|`,
				`| gocov          | 133/141 |   94.33% |`,
				`|   cmd          |     0/3 |    0.00% |`,
				`|     gocov.go   |     0/3 |    0.00% |`,
				`|   internal     | 133/138 |   96.38% |`,
				`|     gocov.go   | 133/138 |   96.38% |`,
				`|----------------|---------|----------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with invalid coverage.out file, invalid column value",
			fsys: fstest.MapFS{
				"go.mod": {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(strings.Join([]string{
					`mode: atomic`,
					`github.com/slavsan/gocov/cmd/gocov.go:9.13,16.22 x 0`,
				}, "\n"))},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"ignore": [`,
					`		"gocov/main.go"`,
					`	]`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "failed to parse coverage file on line 2",
			expectedExitCode: 1,
		},
		{
			title: "with invalid coverage.out file, invalid first line",
			fsys: fstest.MapFS{
				"go.mod": {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(strings.Join([]string{
					`foo: atomic`,
					`github.com/slavsan/gocov/cmd/gocov.go:9.13,16.22 10 0`,
				}, "\n"))},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"ignore": [`,
					`		"gocov/main.go"`,
					`	]`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "invalid coverage file",
			expectedExitCode: 1,
		},
		{
			title: "with invalid .gocov file",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"ignore": [`,
					`		"goc`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "failed to parse .gocov config file: unexpected end of JSON input",
			expectedExitCode: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exiter := &exiterMock{}
			internal.NewCommand(&stdout, &stderr, tc.fsys, tc.config, exiter).Exec(internal.Report, []string{})
			if tc.expectedStdout != stdout.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStdout, stdout.String())
			}
			if tc.expectedStderr != stderr.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStderr, stderr.String())
			}
			if tc.expectedExitCode != exiter.code {
				t.Errorf("exit code does not match\n\texpected:\n`%d`\n\tactual:\n`%d`\n", tc.expectedExitCode, exiter.code)
			}
		})
	}
}

func TestCheckCoverage(t *testing.T) {
	testCases := []struct {
		title            string
		fsys             fs.StatFS
		config           *internal.Config
		expectedStdout   string
		expectedStderr   string
		expectedExitCode int
	}{
		{
			title: "with coverage below threshold",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"threshold": 75.52`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "Coverage check failed: expected to have 75.52 coverage, but got 73.37\n",
			expectedExitCode: 1,
		},
		{
			title: "with coverage above threshold",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"threshold": 23.88`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with missing .gocov file",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "Coverage check failed: missing .gocov file with defined threshold\n",
			expectedExitCode: 1,
		},
		{
			title: "with missing go.mod file",
			fsys: fstest.MapFS{
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "failed to open go.mod file: open go.mod: file does not exist",
			expectedExitCode: 1,
		},
		{
			title: "with invalid go.mod file",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`foo github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "invalid go.mod file",
			expectedExitCode: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exiter := &exiterMock{}
			internal.NewCommand(&stdout, &stderr, tc.fsys, tc.config, exiter).Exec(internal.Check, []string{})
			if tc.expectedStdout != stdout.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStdout, stdout.String())
			}
			if tc.expectedStderr != stderr.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStderr, stderr.String())
			}
			if tc.expectedExitCode != exiter.code {
				t.Errorf("exit code does not match\n\texpected:\n`%d`\n\tactual:\n`%d`\n", tc.expectedExitCode, exiter.code)
			}
		})
	}
}

func TestConfigFile(t *testing.T) {
	testCases := []struct {
		title            string
		fsys             fs.StatFS
		config           *internal.Config
		expectedStdout   string
		expectedStderr   string
		expectedExitCode int
	}{
		{
			title: "with missing config file should return a default config",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`{`,
				`  "threshold": 50,`,
				`  "ignore": [`,
				`  ]`,
				`}`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with defined config file should just return it",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`  "threshold": 75.52`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`{`,
				`  "threshold": 75.52`,
				`}`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exiter := &exiterMock{}
			internal.NewCommand(&stdout, &stderr, tc.fsys, tc.config, exiter).Exec(internal.ConfigFile, []string{})
			if tc.expectedStdout != stdout.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStdout, stdout.String())
			}
			if tc.expectedStderr != stderr.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStderr, stderr.String())
			}
			if tc.expectedExitCode != exiter.code {
				t.Errorf("exit code does not match\n\texpected:\n`%d`\n\tactual:\n`%d`\n", tc.expectedExitCode, exiter.code)
			}
		})
	}
}

func TestInspect(t *testing.T) {
	testCases := []struct {
		title            string
		fsys             fs.StatFS
		config           *internal.Config
		expectedStdout   string
		expectedStderr   string
		expectedExitCode int
	}{
		{
			title: "when printing the entire file",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut3)},
				"cmd/gocov.go": {Data: []byte(strings.Join([]string{
					`package cmd`,
					``,
					`import (`,
					`	"os"`,
					``,
					`	"github.com/slavsan/gocov/internal"`,
					`)`,
					``,
					`func Exec() {`,
					`	var args []string`,
					`	config := &internal.Config{}`,
					`	config.Color = true`,
					``,
					`	command := internal.Report`,
					``,
					`	if len(os.Args) > 1 {`,
					`		switch os.Args[1] {`,
					`		case "check":`,
					`			command = internal.Check`,
					`		case "inspect":`,
					`			command = internal.Inspect`,
					`			if len(os.Args) > 2 {`,
					`				args = append(args, os.Args[2])`,
					`			}`,
					`			//os.Args[1]`,
					`		}`,
					`	}`,
					``,
					`	internal.Exec(`,
					`		command,`,
					`		args,`,
					`		os.Stdout,`,
					`		os.Stderr,`,
					`		os.DirFS("."),`,
					`		config,`,
					`		&internal.ProcessExiter{},`,
					`	)`,
					`}`,
					``,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`1| package cmd`,
				`2| `,
				`3| import (`,
				`4| 	"os"`,
				`5| `,
				`6| 	"github.com/slavsan/gocov/internal"`,
				`7| )`,
				`8| `,
				`9| func Exec() ` + internal.Red + `{`,
				`10| 	var args []string`,
				`11| 	config := &internal.Config{}`,
				`12| 	config.Color = true`,
				`13| `,
				`14| 	command := internal.Report`,
				`15| `,
				`16| 	if len(os.Args) > 1 ` + internal.NoColor + internal.Red + `{`,
				`17| 		switch os.Args[1] ` + internal.NoColor + `{`,
				`18| 		case "check":` + internal.Red,
				`19| 			command = internal.Check` + internal.NoColor,
				`20| 		case "inspect":` + internal.Red,
				`21| 			command = internal.Inspect`,
				`22| 			if len(os.Args) > 2 ` + internal.NoColor + internal.Red + `{`,
				`23| 				args = append(args, os.Args[2])`,
				`24| 			}` + internal.NoColor,
				`25| 			//os.Args[1]`,
				`26| 		}`,
				`27| 	}`,
				`28| `,
				`29| 	` + internal.Red + `internal.Exec(`,
				`30| 		command,`,
				`31| 		args,`,
				`32| 		os.Stdout,`,
				`33| 		os.Stderr,`,
				`34| 		os.DirFS("."),`,
				`35| 		config,`,
				`36| 		&internal.ProcessExiter{},`,
				`37| 	)` + internal.NoColor,
				`38| }`,
				`39| `,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exiter := &exiterMock{}
			internal.NewCommand(&stdout, &stderr, tc.fsys, tc.config, exiter).Exec(internal.Inspect, []string{"gocov/cmd/gocov.go"})
			if tc.expectedStdout != stdout.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStdout, stdout.String())
			}
			if tc.expectedStderr != stderr.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStderr, stderr.String())
			}
			if tc.expectedExitCode != exiter.code {
				t.Errorf("exit code does not match\n\texpected:\n`%d`\n\tactual:\n`%d`\n", tc.expectedExitCode, exiter.code)
			}
		})
	}
}

type exiterMock struct {
	code int
}

func (m *exiterMock) Exit(code int) {
	m.code = code
}
