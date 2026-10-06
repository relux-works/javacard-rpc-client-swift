// Package codegen implements the Swift package backend for javacard-rpc.
//
// Plugin consumes validated pluginapi schemas and returns ordered in-memory
// Package.swift and Swift source files. Parsing, validation and writing are
// facade responsibilities. Stream fields retain the v0.4.5 renderer rejection.
package codegen
