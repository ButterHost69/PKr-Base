package logger

import (
	"fmt"
)

type WorkspaceLogger struct {
	acceptedLevel   LogLevel
	workspacePath   map[string]string // WorkspaceName -> WorkspacePath
	printToTerminal bool
}

func InitWorkspaceLogger() *WorkspaceLogger {
	wl := WorkspaceLogger{
		acceptedLevel:   NoneLevel,
		workspacePath:   map[string]string{},
		printToTerminal: false,
	}
	return &wl
}

func (wl *WorkspaceLogger) SetWorkspacePaths(workspace_paths map[string]string) {
	wl.workspacePath = workspace_paths
}

func (wl *WorkspaceLogger) SetPrintToTerminal(printToTerminal bool) {
	wl.printToTerminal = printToTerminal
}

func (wl *WorkspaceLogger) SetLogLevel(logLevel LogLevel) {
	wl.acceptedLevel = logLevel
}

// [ ] Find a Way to Avoid Frequent Read and Writes... IDK How ? Maybe its not possible ?

func (wl *WorkspaceLogger) Info(workspace_name string, log any){
	slog := fmt.Sprintf("[Info] %v\n", log)
	if wl.printToTerminal {
		fmt.Printf("[Info] %v\n", log)
	}

	if wl.acceptedLevel >= InfoLevel {
		if err := logToFile(wl.workspacePath[workspace_name] + LOGS_PKR_FILE_PATH , slog); err != nil {
			fmt.Println("Could Not Log To File.\nError: ", err)
		}
	}
}

func (wl *WorkspaceLogger) Debug(workspace_name string, log any){
	slog := fmt.Sprintf("[Debug] %v\n", log)
	if wl.printToTerminal {
		fmt.Printf("[Debug] %v\n", log)
	}

	if wl.acceptedLevel >= DebugLevel {
		if err := logToFile(wl.workspacePath[workspace_name] + LOGS_PKR_FILE_PATH, slog); err != nil {
			fmt.Println("Could Not Log To File.\nError: ", err)
		}
	}
}

func (wl *WorkspaceLogger) Critical(workspace_name string, log any){
	slog := fmt.Sprintf("[Critical] %v\n", log)
	if wl.printToTerminal {
		fmt.Printf("[Critical] %v\n", log)
	}

	if wl.acceptedLevel >= CriticalLevel {
		if err := logToFile(wl.workspacePath[workspace_name] + LOGS_PKR_FILE_PATH, slog); err != nil {
			fmt.Println("Could Not Log To File.\nError: ", err)
		}
	}
}