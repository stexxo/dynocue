// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package logging

type NoopLogger struct{}

func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

func (n NoopLogger) Debug(msg string, args ...any) {}
func (n NoopLogger) Info(msg string, args ...any)  {}
func (n NoopLogger) Warn(msg string, args ...any)  {}
func (n NoopLogger) Error(msg string, args ...any) {}
