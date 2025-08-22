# ADR 001: Use PostgreSQL for persistence

## Context
We need a relational database with strong ACID properties, good community support, and open-source licensing.  
Alternatives considered: MySQL, Oracle XE, SQLite.

## Decision
We will use PostgreSQL 16 as the primary data store.  

## Consequences
- Pros: Rich feature set (JSONB, full-text search), strong OSS community.
- Cons: Slightly higher memory footprint than SQLite/MySQL.
- Migration path: Easy scaling with read replicas; potential future use of TimescaleDB extension.