package main

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/common/go/common"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/permission/types"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
)

func cliStringsToInts(cmd *cobra.Command, args []string) {
	cmd.ParseFlags(args)
	if len(args) == 0 {
		Exit(fmt.Errorf("Please enter at least one `<permission>:<value>` pair like `send:0 call:1 create_account:1`"))
	}

	bp := types.ZeroBasePermissions

	for _, a := range args {
		spl := strings.Split(a, ":")
		if len(spl) != 2 {
			Exit(fmt.Errorf("arguments must be like `send:1`, not %s", a))
		}
		name, v := spl[0], spl[1]
		vi := v[0] - '0'
		pf, err := types.PermStringToFlag(name)
		IfExit(err)
		bp.Set(pf, vi > 0)
	}
	printPerms(bp, BitmaskFlag)
}

func coreIntsToStrings(perms, setbits types.PermFlag) map[string]bool {
	m := make(map[string]bool)

	for i := uint(0); i < types.NumPermissions; i++ {
		pf := types.PermFlag(1 << i)
		if pf&setbits > 0 {
			name := types.PermFlagToString(pf)
			m[name] = pf&perms > 0
		}
	}
	return m
}

func cliIntsToStrings(cmd *cobra.Command, args []string) {
	cmd.ParseFlags(args)
	if len(args) != 2 {
		Exit(fmt.Errorf("Please enter PermFlag and SetBit integers"))
	}

	pf, sb := args[0], args[1]
	perms, err := strconv.Atoi(pf)
	IfExit(err)
	setbits, err := strconv.Atoi(sb)
	IfExit(err)

	m := coreIntsToStrings(types.PermFlag(perms), types.PermFlag(setbits))
	printStringPerms(m)
}

func cliBBPB(cmd *cobra.Command, args []string) {
	pf := types.DefaultPermFlags
	printPerms(types.BasePermissions{pf, pf}, BitmaskFlag)

	fmt.Println("")

	m := coreIntsToStrings(pf, pf)
	printStringPerms(m)
}

func cliAll(cmd *cobra.Command, args []string) {
	pf := types.AllPermFlags
	printPerms(types.BasePermissions{pf, pf}, BitmaskFlag)

	fmt.Println("")

	m := coreIntsToStrings(pf, pf)
	printStringPerms(m)
}

func printPerms(bp types.BasePermissions, bits bool) {
	fmt.Println("Perms and SetBit (As Integers)")
	fmt.Printf("%d,%d\n", bp.Perms, bp.SetBit)
	if bits {
		fmt.Println("\nPerms and SetBit (As Bitmasks)")
		fmt.Printf("%b,%b\n", bp.Perms, bp.SetBit)
	}
}

func printStringPerms(m map[string]bool) {
	for i := 0; i < int(types.NumPermissions); i++ {
		permName := types.PermFlagToString(types.PermFlag(1) << uint(i))
		fmt.Printf("%s: %v\n", permName, m[permName])
	}
}
