/*
 * Copyright 1999-2020 Alibaba Group Holding Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package exec

import (
	"context"
	"fmt"
	"path"

	"github.com/chaosblade-io/chaosblade-spec-go/spec"
	"github.com/chaosblade-io/chaosblade-spec-go/util"

	"github.com/chaosblade-io/chaosblade-exec-os/exec/category"
)

const StopProcessBin = "chaos_stopprocess"

type StopProcessActionCommandSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewStopProcessActionCommandSpec() spec.ExpActionCommandSpec {
	return &StopProcessActionCommandSpec{
		spec.BaseExpActionCommandSpec{
			ActionMatchers: []spec.ExpFlagSpec{
				&spec.ExpFlag{
					Name: "process",
					Desc: "Process name",
				},
				&spec.ExpFlag{
					Name: "process-cmd",
					Desc: "Process name in command",
				},
			},
			ActionFlags:    []spec.ExpFlagSpec{},
			ActionExecutor: &StopProcessExecutor{},
			ActionExample: `
# Pause the process that contains the "SimpleHTTPServer" keyword
blade create process stop --process SimpleHTTPServer

# Pause the Java process
blade create process stop --process-cmd java

# Return success even if the process not found
blade create process stop --process demo --ignore-not-found`,
			ActionPrograms:   []string{StopProcessBin},
			ActionCategories: []string{category.SystemProcess},
		},
	}
}

func (*StopProcessActionCommandSpec) Name() string {
	return "stop"
}

func (*StopProcessActionCommandSpec) Aliases() []string {
	return []string{"f"}
}

func (*StopProcessActionCommandSpec) ShortDesc() string {
	return "process fake death"
}

func (s *StopProcessActionCommandSpec) LongDesc() string {
	if s.ActionLongDesc != "" {
		return s.ActionLongDesc
	}
	return "process fake death by process id or process name"
}

type StopProcessExecutor struct {
	channel spec.Channel
}

func (spe *StopProcessExecutor) Name() string {
	return "stop"
}

func (spe *StopProcessExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	if spe.channel == nil {
		return spec.ReturnFail(spec.Code[spec.ServerError], "channel is nil")
	}
	process := model.ActionFlags["process"]
	processCmd := model.ActionFlags["process-cmd"]
	if process == "" && processCmd == "" {
		return spec.ReturnFail(spec.Code[spec.IllegalParameters], "less process matcher")
	}
	ignoreProcessNotFound := model.ActionFlags["ignore-not-found"] == "true"
	flags := fmt.Sprintf("--debug=%t", util.Debug)
	if process != "" {
		flags = fmt.Sprintf(`%s --process "%s"`, flags, process)
	} else if processCmd != "" {
		flags = fmt.Sprintf(`%s --process-cmd "%s"`, flags, processCmd)
	}
	if ignoreProcessNotFound {
		flags = fmt.Sprintf(`%s --ignore-not-found=%t`, flags, ignoreProcessNotFound)
	}

	if _, ok := spec.IsDestroy(ctx); ok {
		return spe.recoverProcess(flags, ctx)
	} else {
		return spe.stopProcess(flags, ctx)
	}
}

func (spe *StopProcessExecutor) stopProcess(flags string, ctx context.Context) *spec.Response {
	flags = fmt.Sprintf(`--start %s`, flags)
	return spe.channel.Run(ctx, path.Join(spe.channel.GetScriptPath(), StopProcessBin), flags)
}

func (spe *StopProcessExecutor) recoverProcess(flags string, ctx context.Context) *spec.Response {
	flags = fmt.Sprintf(`--stop %s`, flags)
	return spe.channel.Run(ctx, path.Join(spe.channel.GetScriptPath(), StopProcessBin), flags)
}

func (spe *StopProcessExecutor) SetChannel(channel spec.Channel) {
	spe.channel = channel
}
