package depgen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewScanner(t *testing.T) {
	s := NewScanner()
	if s == nil {
		t.Error("NewScanner returned nil")
	}
}

func TestScanPython(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `
import os
import sys
import requests
from flask import Flask
from mypackage.submodule import something
from . import relative
import json
from collections import defaultdict
`
	if err := os.WriteFile(pyFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{pyFile}, "python")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	// Should find: requests, flask, mypackage
	// Should NOT find: os, sys, json, collections (stdlib), relative imports
	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp.Package] = true
	}

	if !found["requests"] {
		t.Error("Should find requests")
	}
	if !found["flask"] {
		t.Error("Should find flask")
	}
	if !found["mypackage"] {
		t.Error("Should find mypackage")
	}
	if found["os"] {
		t.Error("Should not find os (stdlib)")
	}
	if found["sys"] {
		t.Error("Should not find sys (stdlib)")
	}
	if found["json"] {
		t.Error("Should not find json (stdlib)")
	}
}

func TestScanPythonComments(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `
# import commented_out
import real_package
# from another import thing
`
	if err := os.WriteFile(pyFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{pyFile}, "python")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	if len(imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(imports))
	}
	if imports[0].Package != "real_package" {
		t.Errorf("Expected real_package, got %s", imports[0].Package)
	}
}

func TestScanJavaScript(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "test.js")

	content := `
import React from 'react';
import { useState } from 'react';
import axios from "axios";
import './styles.css';
import '../utils/helper';
const lodash = require('lodash');
const local = require('./local');
`
	if err := os.WriteFile(jsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{jsFile}, "javascript")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp.Package] = true
	}

	if !found["react"] {
		t.Error("Should find react")
	}
	if !found["axios"] {
		t.Error("Should find axios")
	}
	if !found["lodash"] {
		t.Error("Should find lodash")
	}
	if found["./styles.css"] {
		t.Error("Should not find relative imports")
	}
	if found["./local"] {
		t.Error("Should not find relative requires")
	}
}

func TestScanJavaScriptBuiltins(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "test.js")

	content := `
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const http = require('http');
const nodeFs = require('node:fs');
const nodePath = require('node:path');
import stream from 'stream';
import express from 'express';
import { readFile } from 'node:fs/promises';
`
	if err := os.WriteFile(jsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{jsFile}, "javascript")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp.Package] = true
	}

	if found["fs"] {
		t.Error("Should not find Node.js builtin 'fs'")
	}
	if found["path"] {
		t.Error("Should not find Node.js builtin 'path'")
	}
	if found["crypto"] {
		t.Error("Should not find Node.js builtin 'crypto'")
	}
	if found["http"] {
		t.Error("Should not find Node.js builtin 'http'")
	}
	if found["node:fs"] {
		t.Error("Should not find node:-prefixed builtin 'node:fs'")
	}
	if found["node:path"] {
		t.Error("Should not find node:-prefixed builtin 'node:path'")
	}
	if found["stream"] {
		t.Error("Should not find Node.js builtin 'stream'")
	}
	if !found["express"] {
		t.Error("Should find external package 'express'")
	}
}

func TestScanGo(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	// Scanner only detects imports in import blocks, not single-line imports
	content := `package main

import (
	"fmt"
	"os"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func main() {}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{goFile}, "go")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp.Package] = true
	}

	if !found["github.com/pkg/errors"] {
		t.Error("Should find github.com/pkg/errors")
	}
	if !found["github.com/spf13/cobra"] {
		t.Error("Should find github.com/spf13/cobra")
	}
	if found["fmt"] {
		t.Error("Should not find fmt (stdlib)")
	}
	if found["os"] {
		t.Error("Should not find os (stdlib)")
	}
}

func TestScanRust(t *testing.T) {
	tmpDir := t.TempDir()
	rsFile := filepath.Join(tmpDir, "test.rs")

	content := `
use std::collections::HashMap;
use serde::{Deserialize, Serialize};
use tokio::runtime::Runtime;
use crate::utils::helper;
`
	if err := os.WriteFile(rsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{rsFile}, "rust")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp.Package] = true
	}

	if !found["serde"] {
		t.Error("Should find serde")
	}
	if !found["tokio"] {
		t.Error("Should find tokio")
	}
	if found["std"] {
		t.Error("Should not find std")
	}
	if found["crate"] {
		t.Error("Should not find crate")
	}
}

func TestScanMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "a.py")
	file2 := filepath.Join(tmpDir, "b.py")

	if err := os.WriteFile(file1, []byte("import requests\nimport flask"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("import requests\nimport django"), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{file1, file2}, "python")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	// Should deduplicate requests
	if len(imports) != 3 {
		t.Errorf("Expected 3 unique imports, got %d", len(imports))
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp.Package] = true
	}

	if !found["requests"] || !found["flask"] || !found["django"] {
		t.Error("Missing expected imports")
	}
}

func TestScanByExtension(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with different extensions
	pyFile := filepath.Join(tmpDir, "test.py")
	jsFile := filepath.Join(tmpDir, "test.js")
	tsFile := filepath.Join(tmpDir, "test.ts")
	goFile := filepath.Join(tmpDir, "test.go")
	rsFile := filepath.Join(tmpDir, "test.rs")

	if err := os.WriteFile(pyFile, []byte("import requests"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(jsFile, []byte("import axios from 'axios'"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tsFile, []byte("import axios from 'axios'"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(goFile, []byte(`package main

import (
	"github.com/pkg/errors"
)
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rsFile, []byte("use serde::Serialize;"), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()

	// Test extension-based detection (empty language)
	tests := []struct {
		file     string
		expected string
	}{
		{pyFile, "requests"},
		{jsFile, "axios"},
		{tsFile, "axios"},
		{goFile, "github.com/pkg/errors"},
		{rsFile, "serde"},
	}

	for _, tt := range tests {
		imports, err := s.ScanFiles([]string{tt.file}, "")
		if err != nil {
			t.Errorf("ScanFiles(%s) failed: %v", tt.file, err)
			continue
		}
		if len(imports) == 0 {
			t.Errorf("No imports found in %s", tt.file)
			continue
		}
		if imports[0].Package != tt.expected {
			t.Errorf("ScanFiles(%s) = %s, want %s", tt.file, imports[0].Package, tt.expected)
		}
	}
}

func TestScanNonExistentFile(t *testing.T) {
	s := NewScanner()
	_, err := s.ScanFiles([]string{"/nonexistent/file.py"}, "python")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestScanUnknownExtension(t *testing.T) {
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(txtFile, []byte("some content"), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{txtFile}, "")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	// Should return empty for unknown extension
	if len(imports) != 0 {
		t.Errorf("Expected 0 imports for .txt, got %d", len(imports))
	}
}

func TestImportInfoFields(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `import requests
from flask import Flask`
	if err := os.WriteFile(pyFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	imports, err := s.ScanFiles([]string{pyFile}, "python")
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	for _, imp := range imports {
		if imp.SourceFile != pyFile {
			t.Errorf("SourceFile = %q, want %q", imp.SourceFile, pyFile)
		}
		if imp.LineNumber < 1 {
			t.Error("LineNumber should be >= 1")
		}
	}
}
