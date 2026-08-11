package cobra

import (
	"bytes"
	"testing"
)

func TestResetFlags(t *testing.T) {
	var host string

	rootCmd := &Command{
		Use: "root",
		Run: func(cmd *Command, args []string) {},
	}
	rootCmd.PersistentFlags().StringVar(&host, "host", "localhost", "host address")

	subCmd := &Command{
		Use: "sub",
		Run: func(cmd *Command, args []string) {},
	}
	rootCmd.AddCommand(subCmd)

	// 1. Execute with --host remotehost
	rootCmd.SetArgs([]string{"sub", "--host", "remotehost"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if host != "remotehost" {
		t.Errorf("expected host to be 'remotehost', got '%s'", host)
	}

	// 2. Execute again with no arguments — NO manual ResetFlags call. The
	// automatic reset inside Execute() must return host to its default.
	rootCmd.SetArgs([]string{"sub"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if host != "localhost" {
		t.Errorf("expected host to be reset to 'localhost' by Execute, got '%s'", host)
	}
}

func TestExecutePreservesProgrammaticFlagOnFirstRun(t *testing.T) {
	var port string

	rootCmd := NewCommand()
	rootCmd.Run = func(cmd *Command, args []string) {}

	rootCmd.PersistentFlags().StringVar(&port, "port", "8080", "port")
	// Programmatic set before first execution must survive the first run.
	if err := rootCmd.PersistentFlags().Set("port", "9090"); err != nil {
		t.Fatalf("set: %v", err)
	}

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != "9090" {
		t.Errorf("expected programmatic port 9090 on first run, got %q", port)
	}

	// Second execution resets to default.
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != "8080" {
		t.Errorf("expected port reset to 8080 on second run, got %q", port)
	}
}

func TestResetFlagsSlice(t *testing.T) {
	var hosts []string

	rootCmd := &Command{
		Use: "root",
		Run: func(cmd *Command, args []string) {},
	}
	rootCmd.PersistentFlags().StringSliceVar(&hosts, "hosts", []string{"localhost"}, "host addresses")

	subCmd := &Command{
		Use: "sub",
		Run: func(cmd *Command, args []string) {},
	}
	rootCmd.AddCommand(subCmd)

	// 1. Execute with --hosts remote1,remote2
	rootCmd.SetArgs([]string{"sub", "--hosts", "remote1,remote2"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hosts) != 2 || hosts[0] != "remote1" || hosts[1] != "remote2" {
		t.Errorf("expected hosts to be [remote1, remote2], got %v", hosts)
	}

	// 2. Execute again with no arguments — automatic reset inside Execute()
	// must return hosts to its default.
	rootCmd.SetArgs([]string{"sub"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hosts) != 1 || hosts[0] != "localhost" {
		t.Errorf("expected hosts to be reset to [localhost] by Execute, got %v", hosts)
	}
}
