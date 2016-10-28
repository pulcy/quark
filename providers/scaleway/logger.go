// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scaleway

import (
	"net/http"

	logging "github.com/op/go-logging"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

type scalewayLogger struct {
	log *logging.Logger
}

func NewScalewayLogger(log *logging.Logger) api.Logger {
	return &scalewayLogger{
		log: log,
	}
}

func (l *scalewayLogger) LogHTTP(r *http.Request) {
	l.log.Debugf("%s %s\n", r.Method, r.URL.RawPath)
}

func (l *scalewayLogger) Fatalf(format string, v ...interface{}) {
	l.log.Fatalf(format, v...)
}

func (l *scalewayLogger) Debugf(format string, v ...interface{}) {
	l.log.Debugf(format, v...)
}

func (l *scalewayLogger) Infof(format string, v ...interface{}) {
	l.log.Infof(format, v...)
}

func (l *scalewayLogger) Warnf(format string, v ...interface{}) {
	l.log.Warningf(format, v...)
}
