package feedback

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	feedbackCmd = &cobra.Command{
		Use:   "feedback",
		Short: "Request for customer feedback on specified topics",
		RunE:  getFeedback,
	}
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
)

var topicsDB = map[string]string{
	"observability-container-logging_1":  "[Example] Your workload is using DeploymentConfig, which does not follow the recommendation of using Deployments or StatefulSets. Why are they used?",
	"observability-termination-policy_1": "[Example] Your application uses a port (1001) that is restricted. Could you use another one?",
	"observability-termination-policy_2": "[Example] An SSH daemon has been found running in a Pod? That's not recommended. Why is it needed?",
}

func provideFeedback() bool {
	prompt := promptui.Select{
		Label:        "Would you like to provide feedback?",
		Items:        []string{"Yes", "No"},
		HideHelp:     true,
		HideSelected: true,
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	return result == "Yes"
}

func getFeedbackTopics() []string {
	data, err := os.ReadFile("./cnf-certification-test/claim.json")
	if err != nil {
		log.Fatalf("Error reading claim file: %v", err)
		return nil
	}

	r, _ := regexp.Compile(`\[CQ#([^\]]+)`)
	matches := r.FindAllStringSubmatch(string(data), -1)

	m := make(map[string]bool)
	for _, topic := range matches {
		m[topic[1]] = true
	}

	topics := make([]string, 0, len(m))
	for topic := range m {
		if _, ok := topicsDB[topic]; ok {
			topics = append(topics, topic)
		}
	}

	return topics
}

func prependCommentSign(s string, times int) string {
	commentSign := strings.Repeat("#", times)
	lines := strings.Split(s, "\n")
	commentedLines := make([]string, len(lines))
	for i, line := range lines {
		if line != "" {
			commentedLines[i] = commentSign + " " + line
		}
	}

	return strings.Join(commentedLines, "\n")
}

func requestFeedback(question string) (string, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %v", err)
	}

	tmpFile.WriteString("\n" + prependCommentSign(question, 1))
	tmpFile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		path, err := exec.LookPath("vi")
		if err != nil {
			return "", fmt.Errorf("vi is not available: %v", err)
		}
		editor = path
	}
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return "", fmt.Errorf("error running command %s: %v", cmd.String(), err)
	}
	err = cmd.Wait()
	if err != nil {
		return "", fmt.Errorf("error waiting for command %s: %v", cmd.String(), err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %v", tmpFile.Name(), err)
	}

	return strings.Split(string(data), "#")[0], nil
}

func getFeedback(cmd *cobra.Command, args []string) error {
	answersFile, err := os.Create("customer-feedback.md")
	if err != nil {
		return fmt.Errorf("error creating answers file: %v", err)
	}
	defer answersFile.Close()

	topics := getFeedbackTopics()
	fmt.Printf("Number of topics that require customer feedback: %d\n", len(topics))

	for _, topic := range topics {
		topicQuestion := topicsDB[topic]
		fmt.Printf("\n[%s]\n%s\n", topic, topicQuestion)
		var topicAnswer string
		if !provideFeedback() {
			topicAnswer = "No answer provided.\n"
			fmt.Printf("%s%s%s", colorRed, "No answer provided.\n", colorReset)
		} else {
			topicAnswer, err = requestFeedback(topicQuestion)
			if err != nil {
				return fmt.Errorf("could not get feedback, err: %v", err)
			}
			fmt.Printf("%s%s%s", colorGreen, "Answer saved.\n", colorReset)
		}

		answersFile.WriteString(prependCommentSign("["+topic+"]\n"+topicQuestion+"\n", 3))
		answersFile.WriteString(topicAnswer + "\n")
	}

	return nil
}

func NewCommand() *cobra.Command {
	return feedbackCmd
}
