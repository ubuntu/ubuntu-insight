package sysinfo_test

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubuntu/ubuntu-insights/internal/collector/sysinfo"
	"github.com/ubuntu/ubuntu-insights/internal/testutils"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
	}{
		"Instantiate a sys info manager": {},
	}
	for name := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := sysinfo.New(sysinfo.WithRoot("/myspecialroot"))

			require.NotEmpty(t, s, "sysinfo manager has custom fields")
		})
	}
}

func TestCollect(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		root         string
		cpuInfo      string
		blkInfo      string
		screenInfo   string
		missingFiles []string

		logs    map[slog.Level]uint
		wantErr bool
	}{
		"Regular hardware information": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "regular",
			screenInfo: "regular",
		},

		"Missing Product information": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "regular",
			screenInfo: "regular",
			missingFiles: []string{
				"sys/class/dmi/id/product_family",
				"sys/class/dmi/id/product_name",
				"sys/class/dmi/id/sys_vendor",
			},

			logs: map[slog.Level]uint{
				slog.LevelWarn: 3,
			},
		},

		"Missing CPU information": {
			root:       "regular",
			cpuInfo:    "missing",
			blkInfo:    "regular",
			screenInfo: "regular",

			logs: map[slog.Level]uint{
				slog.LevelWarn: 1,
			},
		},

		"Garbage CPU information is empty": {
			root:       "regular",
			cpuInfo:    "garbage",
			blkInfo:    "regular",
			screenInfo: "regular",

			logs: map[slog.Level]uint{
				slog.LevelWarn: 1,
			},
		},

		"Missing GPUs": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "regular",
			screenInfo: "regular",
			missingFiles: []string{
				"sys/class/drm/card0",
				"sys/class/drm/card1",
			},

			logs: map[slog.Level]uint{
				slog.LevelWarn: 1,
			},
		},

		"Missing GPU information": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "regular",
			screenInfo: "regular",
			missingFiles: []string{
				"sys/class/drm/c0/d0/driver",
				"sys/class/drm/c0/d0/label",
				"sys/class/drm/c0/d0/vendor",
			},

			logs: map[slog.Level]uint{
				slog.LevelWarn: 3,
			},
		},

		"Missing Memory information": {
			root:         "regular",
			cpuInfo:      "regular",
			blkInfo:      "regular",
			screenInfo:   "regular",
			missingFiles: []string{"proc/meminfo"},

			logs: map[slog.Level]uint{
				slog.LevelWarn: 1,
			},
		},

		"Missing Block information": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "missing",
			screenInfo: "regular",

			logs: map[slog.Level]uint{
				slog.LevelWarn: 1,
			},
		},

		"Garbage Block information is empty": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "garbage",
			screenInfo: "regular",

			logs: map[slog.Level]uint{
				slog.LevelWarn: 1,
			},
		},

		"Missing Screen information": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "regular",
			screenInfo: "",

			logs: map[slog.Level]uint{
				slog.LevelWarn: 1,
			},
		},

		"Garbage Screen information is empty": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "regular",
			screenInfo: "garbage",

			logs: map[slog.Level]uint{
				slog.LevelWarn: 1,
			},
		},

		"Missing Screen refresh information is empty": {
			root:       "regular",
			cpuInfo:    "regular",
			blkInfo:    "regular",
			screenInfo: "missing refresh",

			logs: map[slog.Level]uint{
				slog.LevelWarn: 2,
			},
		},

		"Missing hardware information is empty": {
			root:       "withoutinfo",
			cpuInfo:    "",
			blkInfo:    "",
			screenInfo: "",
			logs: map[slog.Level]uint{
				slog.LevelWarn: 8,
			},
		},

		"Empty hardware information is empty": {
			root:       "empty",
			cpuInfo:    "",
			blkInfo:    "",
			screenInfo: "",
			logs: map[slog.Level]uint{
				slog.LevelWarn: 4,
			},
		},

		"Garbage hardware information is sane": {
			root:       "garbage",
			cpuInfo:    "garbage",
			blkInfo:    "garbage",
			screenInfo: "garbage",
			logs: map[slog.Level]uint{
				slog.LevelWarn: 18,
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			err := testutils.CopyDir("testdata/linuxfs", tmp)
			if err != nil {
				fmt.Printf("Setup: failed to copy testdata directory: %v\n", err)
				t.FailNow()
			}

			root := filepath.Join(tmp, tc.root)
			for _, f := range tc.missingFiles {
				err := os.Remove(filepath.Join(root, f))
				if err != nil {
					fmt.Printf("Setup: failed to remove file %s: %v\n", f, err)
					t.FailNow()
				}
			}

			l := testutils.NewMockHandler()

			options := []sysinfo.Options{
				sysinfo.WithRoot(root),
				sysinfo.WithLogger(&l),
			}

			cover := fmt.Sprintf("GOCOVERDIR=%s", os.Getenv("GOCOVERDIR"))
			if tc.cpuInfo != "-" {
				cmdArgs := []string{"env", "GO_WANT_HELPER_PROCESS=1", cover, os.Args[0], "-test.run=TestMockCPUList", "--"}
				cmdArgs = append(cmdArgs, tc.cpuInfo)
				options = append(options, sysinfo.WithCPUInfo(cmdArgs))
			}

			if tc.blkInfo != "-" {
				cmdArgs := []string{"env", "GO_WANT_HELPER_PROCESS=1", cover, os.Args[0], "-test.run=TestMockBlkList", "--"}
				cmdArgs = append(cmdArgs, tc.blkInfo)
				options = append(options, sysinfo.WithBlkInfo(cmdArgs))
			}

			if tc.blkInfo != "-" {
				cmdArgs := []string{"env", "GO_WANT_HELPER_PROCESS=1", cover, os.Args[0], "-test.run=TestMockScreenList", "--"}
				cmdArgs = append(cmdArgs, tc.screenInfo)
				options = append(options, sysinfo.WithScreenInfo(cmdArgs))
			}

			s := sysinfo.New(options...)

			got, err := s.Collect()
			if tc.wantErr {
				require.Error(t, err, "Collect should return an error and didn’t")
				return
			}
			require.NoError(t, err, "Collect should not return an error")

			want := testutils.LoadWithUpdateFromGoldenYAML(t, got)
			assert.Equal(t, want, got, "Collect should return expected sys information")

			if !l.AssertLevels(t, tc.logs) {
				for _, call := range l.HandleCalls {
					fmt.Printf("Logged %v: %s\n", call.Level, call.Message)
				}
			}
		})
	}
}

func TestMockCPUList(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] != "--" {
			args = args[1:]
			continue
		}
		args = args[1:]
		break
	}

	switch args[0] {
	case "exit 1":
		fmt.Fprint(os.Stderr, "Error requested in Mock cpulist")
		os.Exit(1)
	case "regular":
		fmt.Println(`{
   "lscpu": [
      {
         "field": "Architecture:",
         "data": "x86_64",
         "children": [
            {
               "field": "CPU op-mode(s):",
               "data": "32-bit, 64-bit"
            },{
               "field": "Address sizes:",
               "data": "39 bits physical, 48 bits virtual"
            },{
               "field": "Byte Order:",
               "data": "Little Endian"
            }
         ]
      },{
         "field": "CPU(s):",
         "data": "12",
         "children": [
            {
               "field": "On-line CPU(s) list:",
               "data": "0-11"
            }
         ]
      },{
         "field": "Vendor ID:",
         "data": "GenuineIntel",
         "children": [
            {
               "field": "Model name:",
               "data": "Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz",
               "children": [
                  {
                     "field": "CPU family:",
                     "data": "6"
                  },{
                     "field": "Model:",
                     "data": "158"
                  },{
                     "field": "Thread(s) per core:",
                     "data": "2"
                  },{
                     "field": "Core(s) per socket:",
                     "data": "6"
                  },{
                     "field": "Socket(s):",
                     "data": "1"
                  },{
                     "field": "Stepping:",
                     "data": "10"
                  },{
                     "field": "CPU(s) scaling MHz:",
                     "data": "22%"
                  },{
                     "field": "CPU max MHz:",
                     "data": "4100.0000"
                  },{
                     "field": "CPU min MHz:",
                     "data": "800.0000"
                  },{
                     "field": "BogoMIPS:",
                     "data": "4399.99"
                  },{
                     "field": "Flags:",
                     "data": "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc cpuid aperfmperf pni pclmulqdq dtes64 monitor ds_cpl vmx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch cpuid_fault epb pti ssbd ibrs ibpb stibp tpr_shadow flexpriority ept vpid ept_ad fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid mpx rdseed adx smap clflushopt intel_pt xsaveopt xsavec xgetbv1 xsaves dtherm ida arat pln pts hwp hwp_notify hwp_act_window hwp_epp vnmi md_clear flush_l1d arch_capabilities"
                  }
               ]
            }
         ]
      },{
         "field": "Caches (sum of all):",
         "data": null,
         "children": [
            {
               "field": "L1d:",
               "data": "192 KiB (6 instances)"
            },{
               "field": "L1i:",
               "data": "192 KiB (6 instances)"
            },{
               "field": "L2:",
               "data": "1.5 MiB (6 instances)"
            },{
               "field": "L3:",
               "data": "9 MiB (1 instance)"
            }
         ]
      }
   ]
}`)
	case "garbage":
		fmt.Println("-100cpus, 10 sockets, 20 cores per socket, 400 threads per core, nebula computing enabled.")
	case "":
		fallthrough
	case "missing":
		os.Exit(0)
	}
}

func TestMockBlkList(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] != "--" {
			args = args[1:]
			continue
		}
		args = args[1:]
		break
	}

	switch args[0] {
	case "exit 1":
		fmt.Fprint(os.Stderr, "Error requested in Mock lsblk")
		os.Exit(1)
	case "regular":
		fmt.Println(`{
   "blockdevices": [
      {
         "name": "loop0",
         "size": "4K",
         "type": "loop"
      },{
         "name": "loop1",
         "size": "9.5M",
         "type": "loop"
      },{
         "name": "sda",
         "size": "931.5G",
         "type": "disk",
         "children": [
            {
               "name": "sda1",
               "size": "1G",
               "type": "part"
            },{
               "name": "sda2",
               "size": "2G",
               "type": "part"
            },{
               "name": "sda3",
               "size": "928.5G",
               "type": "part",
               "children": [
                  {
                     "name": "dm_crypt-0",
                     "size": "928.4G",
                     "type": "crypt",
                     "children": [
                        {
                           "name": "ubuntu--vg-ubuntu--lv",
                           "size": "928.4G",
                           "type": "lvm"
                        }
                     ]
                  }
               ]
            }
         ]
      }
   ]
}`)
	case "garbage":
		fmt.Println(`my ssd is broken :(
I should get a new one.`)
	case "":
		fallthrough
	case "missing":
		os.Exit(0)
	}
}

func TestMockScreenList(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] != "--" {
			args = args[1:]
			continue
		}
		args = args[1:]
		break
	}

	switch args[0] {
	// TODO: case with non matching regexp
	case "exit 1":
		fmt.Fprint(os.Stderr, "Error requested in Mock lsblk")
		os.Exit(1)
	case "regular":
		fmt.Println(`Screen 0: minimum 8 x 8, current 6912 x 2160, maximum 32767 x 32767
HDMI-0 connected primary 3840x2160+3072+0 (normal left inverted right x axis y axis) 598mm x 336mm
   1920x1080     60.00*+ 100.00    84.90    74.97    59.94    50.00  
   1680x1050     59.95  
   1440x900      59.89  
   1280x1024     75.02    60.02  
   1280x960      60.00  
   1280x800      59.81  
   1280x720      60.00    59.94    50.00  
   1152x864      75.00  
   1024x768      75.03    70.07    60.00  
   800x600       75.00    72.19    60.32    56.25  
   720x576       50.00  
   720x480       59.94  
   640x480       75.00    72.81    59.94    59.93  
DP-0 disconnected (normal left inverted right x axis y axis)
DP-1 disconnected (normal left inverted right x axis y axis)
eDP-1-1 connected 3072x1728+0+432 (normal left inverted right x axis y axis) 344mm x 193mm
   1920x1080     60.03*+  60.03    40.02  
   1680x1050     60.03  
   1400x1050     60.03  
   1600x900      60.03  
   1280x1024     60.03  
   1400x900      60.03  
   1280x960      60.03  
   1440x810      60.03  
   1368x768      60.03  
   1280x800      60.03  
   1280x720      60.03  
   1024x768      60.03  
   960x720       60.03  
   928x696       60.03  
   896x672       60.03  
   1024x576      60.03  
   960x600       60.03  
   960x540       60.03  
   800x600       60.03  
   840x525       60.03  
   864x486       60.03  
   700x525       60.03  
   800x450       60.03  
   640x512       60.03  
   700x450       60.03  
   640x480       60.03  
   720x405       60.03  
   684x384       60.03  
   640x360       60.03  
   512x384       60.03  
   512x288       60.03  
   480x270       60.03  
   400x300       60.03  
   432x243       60.03  
   320x240       60.03  
   360x202       60.03  
   320x180       60.03  
DP-1-1 disconnected (normal left inverted right x axis y axis)
HDMI-1-1 disconnected (normal left inverted right x axis y axis)`)
	case "garbage":
		fmt.Println(`HDMI connected! (inverted Down down Up up right left right left a b) primary
If only there were an arcade somewhere...`)
	case "missing refresh":
		fmt.Println(`Screen 0: minimum 8 x 8, current 6912 x 2160, maximum 32767 x 32767
HDMI-0 connected primary 3840x2160+3072+0 (normal left inverted right x axis y axis) 598mm x 336mm
   1920x1080     60.00+ 100.00    84.90`)
	case "":
		fallthrough
	case "missing":
		os.Exit(0)
	}
}
