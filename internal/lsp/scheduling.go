package lsp

import (
	"path/filepath"
	"time"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) scheduleWorkspaceAnalysis(notify glsp.NotifyFunc, document Document) {
	key := s.workspaceScheduleKey(document)
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	if scheduled, ok := s.scheduled[key]; ok {
		scheduled.timer.Stop()
	}
	serial := s.nextScheduleSerialLocked(key)
	s.scheduled[key] = scheduledAnalysis{serial: serial}
	timer := time.AfterFunc(s.debounce, func() {
		if !s.isScheduled(key, serial) {
			return
		}
		current := document
		if open, ok := s.documents.Get(document.URI); ok {
			if document.Version != 0 && (open.Version != document.Version || open.Hash != document.Hash) {
				return
			}
			current = open
		}
		if !s.publishWorkspaceDiagnostics(notify, current) {
			version := current.Version
			if version == 0 {
				version = -1
			}
			publishDiagnostics(notify, protocol.DocumentUri(current.URI), version, s.analyzeDocument(current))
		}
		s.clearScheduled(key, serial)
	})
	s.scheduled[key] = scheduledAnalysis{timer: timer, serial: serial}
}

func (s *Server) workspaceScheduleKey(document Document) string {
	workspaceDocument, ok := s.loadWorkspace(document)
	if !ok {
		return document.URI
	}
	for _, file := range workspaceDocument.Files {
		if filepath.Base(file.Path) == "main.dsl" {
			return "workspace:" + filepath.Dir(file.Path)
		}
	}
	return document.URI
}

func (s *Server) cancelScheduled(uri string) {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	s.cancelScheduledLocked(uri)
}

func (s *Server) cancelScheduledLocked(uri string) {
	if scheduled, ok := s.scheduled[uri]; ok {
		if scheduled.timer != nil {
			scheduled.timer.Stop()
		}
		delete(s.scheduled, uri)
	}
	s.serials[uri]++
}

func (s *Server) cancelAllScheduled() {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	for uri := range s.scheduled {
		s.cancelScheduledLocked(uri)
	}
}

func (s *Server) nextScheduleSerialLocked(uri string) uint64 {
	s.serials[uri]++
	return s.serials[uri]
}

func (s *Server) isScheduled(uri string, serial uint64) bool {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	scheduled, ok := s.scheduled[uri]
	return ok && scheduled.serial == serial && s.serials[uri] == serial
}

func (s *Server) clearScheduled(uri string, serial uint64) {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	if scheduled, ok := s.scheduled[uri]; ok && scheduled.serial == serial {
		delete(s.scheduled, uri)
	}
}
