# Go DNS Server

A simple DNS server implemented from scratch in Go. This project demonstrates building a basic DNS server handling both UDP and TCP queries, including support for large TCP responses, proper partial reads, client disconnect handling, and connection timeouts.

## Features

- UDP and TCP DNS server support
- Handles DNS queries of any length over TCP
- Proper handling of partial reads and client disconnects
- Read and write timeouts to prevent stuck connections
- Simple example response (placeholder for integrating a DNS resolver)
- Written entirely in Go using the `net` package
