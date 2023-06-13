package ebpf

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aquasecurity/tracee/pkg/events"
)

func Test_getTailCalls(t *testing.T) {
	testCases := []struct {
		name              string
		events            map[events.ID]eventConfig
		expectedTailCalls []*events.TailCall
		expectedErr       error
	}{
		{
			name: "happy path - some direct syscalls and syscall requiring events",
			events: map[events.ID]eventConfig{
				events.Ptrace:           {submit: ^uint64(0), emit: ^uint64(0)},
				events.ClockSettime:     {submit: ^uint64(0), emit: ^uint64(0)},
				events.SecurityFileOpen: {submit: ^uint64(0), emit: ^uint64(0)},
				events.MemProtAlert:     {submit: ^uint64(0), emit: ^uint64(0)},
				events.SocketDup:        {submit: ^uint64(0), emit: ^uint64(0)},
			},
			expectedTailCalls: []*events.TailCall{
				events.NewTailCall(
					"sys_exit_tails",
					"sys_dup_exit_tail",
					[]uint32{
						uint32(events.Dup),
						uint32(events.Dup2),
						uint32(events.Dup3),
					},
				),
				events.NewTailCall(
					"sys_enter_init_tail",
					"sys_enter_init",
					[]uint32{
						uint32(events.Dup),
						uint32(events.Dup2),
						uint32(events.Dup3),
					},
				),
				events.NewTailCall(
					"sys_exit_init_tail",
					"sys_exit_init",
					[]uint32{
						uint32(events.Dup),
						uint32(events.Dup2),
						uint32(events.Dup3),
					},
				),
				events.NewTailCall(
					"sys_enter_init_tail",
					"sys_enter_init",
					[]uint32{
						uint32(events.Open),
						uint32(events.Openat),
						uint32(events.Openat2),
						uint32(events.OpenByHandleAt),
						uint32(events.Execve),
						uint32(events.Execveat),
					},
				),
				events.NewTailCall(
					"sys_enter_init_tail",
					"sys_enter_init",
					[]uint32{
						uint32(events.Mmap),
						uint32(events.Mprotect),
						uint32(events.PkeyMprotect),
					},
				),
				events.NewTailCall(
					"sys_enter_init_tail",
					"sys_enter_init",
					[]uint32{
						uint32(events.Ptrace),
						uint32(events.ClockSettime),
					},
				),
				events.NewTailCall(
					"sys_enter_submit_tail",
					"sys_enter_submit",
					[]uint32{
						uint32(events.Ptrace),
						uint32(events.ClockSettime),
					},
				),
				events.NewTailCall(
					"sys_exit_init_tail",
					"sys_exit_init",
					[]uint32{
						uint32(events.Ptrace),
						uint32(events.ClockSettime),
					},
				),
				events.NewTailCall(
					"sys_exit_submit_tail",
					"sys_exit_submit",
					[]uint32{
						uint32(events.Ptrace),
						uint32(events.ClockSettime),
					},
				),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name,
			func(t *testing.T) {
				tailCalls, err := getTailCalls(tc.events)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
					return
				}
				require.NoError(t, err)
				for i := 0; i < len(tailCalls); i++ {
					found := false
					for j := 0; j < len(tc.expectedTailCalls); j++ {
						if tailCalls[i].GetMapName() != tc.expectedTailCalls[j].GetMapName() {
							continue
						}
						if tailCalls[i].GetProgName() != tc.expectedTailCalls[j].GetProgName() {
							continue
						}
						if !reflect.DeepEqual(tailCalls[i].GetMapIndexes(),
							tc.expectedTailCalls[j].GetMapIndexes(),
						) {
							continue
						}
						found = true
					}
					assert.True(t, found)
				}
			},
		)
	}
}
