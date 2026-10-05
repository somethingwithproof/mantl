package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/framework"
	"github.com/thomasvincent/mantl/pkg/runtimeevents"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

var complianceNamespace string
var complianceCmd = &cobra.Command{Use: "compliance", Short: "Inspect compliance coverage, findings, and evidence"}
var queryCompliance = func(ctx context.Context, kind, name, namespace string) ([]byte, error) {
	args := []string{"get", kind}
	if name != "" {
		args = append(args, name)
	}
	args = append(args, "-n", namespace, "-o", "json")
	data, err := exec.CommandContext(ctx, "kubectl", kubectlArgs(args...)...).Output()
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", kind, err)
	}
	return data, nil
}
var complianceStatusCmd = &cobra.Command{Use: "status", Short: "Print profiles, findings, and audit freshness as JSON", RunE: func(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
	defer cancel()
	result := map[string]json.RawMessage{}
	for _, kind := range []string{"complianceprofiles", "findings", "complianceaudits", "auditruns", "controlevaluations", "complianceexceptions"} {
		data, err := queryCompliance(ctx, kind, "", complianceNamespace)
		if err != nil {
			return err
		}
		if !json.Valid(data) {
			return fmt.Errorf("invalid %s response", kind)
		}
		result[kind] = data
	}
	return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
}}
var exportOutput string
var exportRun bool
var exportCmd = &cobra.Command{Use: "export AUDIT", Short: "Download and verify an evidence archive", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
	defer cancel()
	kind := "complianceaudit"
	if exportRun {
		kind = "auditrun"
	}
	data, err := queryCompliance(ctx, kind, args[0], complianceNamespace)
	if err != nil {
		return err
	}
	var audit api.ComplianceAudit
	if err := json.Unmarshal(data, &audit); err != nil {
		return err
	}
	u, err := url.Parse(audit.Status.EvidenceURI)
	if err != nil || u.Scheme != "s3" || u.Host == "" || u.Query().Get("versionId") == "" || audit.Status.ManifestHash == "" {
		return fmt.Errorf("audit has no versioned, hashed evidence manifest")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	store := &evidence.S3Store{Client: s3.NewFromConfig(cfg), Bucket: u.Host}
	version := u.Query().Get("versionId")
	u.RawQuery = ""
	ref := evidence.ObjectRef{URI: u.String(), Version: version, Hash: audit.Status.ManifestHash}
	temp, err := os.CreateTemp(filepath.Dir(exportOutput), ".mantl-evidence-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }()
	defer func() { _ = temp.Close() }()
	if err := evidence.Export(ctx, store, ref, temp); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(temp.Name(), exportOutput); err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), "Verified evidence written to", exportOutput)
	return err
}}
var falcoFrameworkDir string
var falcoMappings string
var falcoCmd = &cobra.Command{Use: "ingest-falco EVENT.json", Short: "Convert a Falco event into a mapped Finding manifest for reviewed ingestion", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	if !cmd.Flags().Changed("framework-dir") {
		falcoFrameworkDir = filepath.Join(sourceDir, "compliance/frameworks")
	}
	if !cmd.Flags().Changed("mappings") {
		falcoMappings = filepath.Join(sourceDir, "compliance/integration/falco-mappings.yaml")
	}
	info, err := os.Stat(args[0])
	if err != nil {
		return err
	}
	if info.Size() > 1024*1024 {
		return fmt.Errorf("falco event exceeds 1 MiB")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	signal, err := runtimeevents.Parse(data)
	if err != nil {
		return err
	}
	mappingData, err := os.ReadFile(falcoMappings)
	if err != nil {
		return err
	}
	var mappings map[string]string
	if err := yaml.UnmarshalStrict(mappingData, &mappings); err != nil {
		return err
	}
	control := mappings[signal.Rule]
	if control == "" {
		return fmt.Errorf("falco rule has no reviewed control mapping")
	}
	fw, err := framework.Load(falcoFrameworkDir, "soc2", "")
	if err != nil {
		return err
	}
	found := false
	for _, c := range fw.Controls {
		if c.ID == control {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("falco mapping references an unknown SOC2 control")
	}
	finding := api.Finding{TypeMeta: metav1.TypeMeta{APIVersion: "compliance.mantl.io/v1alpha1", Kind: "Finding"}, ObjectMeta: metav1.ObjectMeta{Name: signal.ID, Namespace: complianceNamespace, Labels: map[string]string{"mantl.io/source": "falco"}, Annotations: map[string]string{"mantl.io/observed-at": signal.At.Format(time.RFC3339Nano)}}, Spec: api.FindingSpec{ID: signal.ID, ControlID: control, Framework: "soc2", Severity: signal.Severity, Resource: "Pod/" + signal.Namespace + "/" + signal.Pod, Message: signal.Rule, Status: "fail"}}
	payload, err := yaml.Marshal(finding)
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(payload)
	return err
}}
var remediationGuidance string
var remediateCmd = &cobra.Command{Use: "remediation", Short: "Print GitOps remediation suggestions for unresolved findings", RunE: func(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
	defer cancel()
	data, err := queryCompliance(ctx, "findings", "", complianceNamespace)
	if err != nil {
		return err
	}
	var findings api.FindingList
	if err := json.Unmarshal(data, &findings); err != nil {
		return err
	}
	type suggestion struct {
		Finding  string   `json:"finding"`
		Resource string   `json:"resource"`
		Control  string   `json:"control"`
		Actions  []string `json:"actions"`
	}
	if !cmd.Flags().Changed("guidance") {
		remediationGuidance = filepath.Join(sourceDir, "compliance/remediation.yaml")
	}
	guidanceData, err := os.ReadFile(remediationGuidance)
	if err != nil {
		return err
	}
	var guidance map[string][]string
	if err := yaml.UnmarshalStrict(guidanceData, &guidance); err != nil {
		return err
	}
	out := []suggestion{}
	for _, f := range findings.Items {
		if f.Spec.Status != "fail" {
			continue
		}
		actions := []string{}
		for _, id := range strings.Split(f.Spec.ControlID, ",") {
			actions = append(actions, guidance[id]...)
		}
		if len(actions) == 0 {
			actions = append(actions, "Review the failing policy rule and correct the resource in its GitOps source.")
		}
		actions = append(actions, "Open a reviewed GitOps pull request, then verify the next PolicyReport and audit before closing.")
		out = append(out, suggestion{f.Name, f.Spec.Resource, f.Spec.ControlID, actions})
	}
	return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
}}

func init() {
	rootCmd.AddCommand(complianceCmd)
	complianceCmd.PersistentFlags().StringVar(&complianceNamespace, "namespace", "mantl-compliance-system", "Namespace containing compliance metadata")
	complianceCmd.AddCommand(complianceStatusCmd, exportCmd, falcoCmd, remediateCmd)
	remediateCmd.Flags().StringVar(&remediationGuidance, "guidance", "compliance/remediation.yaml", "Reviewed control remediation guidance")
	exportCmd.Flags().BoolVar(&exportRun, "run", false, "Export a historical AuditRun instead of the audit latest pointer")
	exportCmd.Flags().StringVar(&exportOutput, "output", "evidence.tar.gz", "Verified archive output")
	falcoCmd.Flags().StringVar(&falcoFrameworkDir, "framework-dir", "compliance/frameworks", "Packaged framework directory")
	falcoCmd.Flags().StringVar(&falcoMappings, "mappings", "compliance/integration/falco-mappings.yaml", "Reviewed rule-to-control mapping")
}
