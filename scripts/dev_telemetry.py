#!/usr/bin/env python3
"""Dev-session telemetry helper for UniversalOps.

Queries the local SQLite DB (data/universalops.db) to help the agent follow
along with the running app: recent log entries, error/warn summary, and
recent metric inserts.

Usage:
    python scripts/dev_telemetry.py schema                 # dump table list + columns
    python scripts/dev_telemetry.py logs   [--limit N]    # recent log entries (N=25)
    python scripts/dev_telemetry.py errors [--limit N]    # ERROR/WARN entries
    python scripts/dev_telemetry.py metrics [--limit N]   # recent metric inserts
    python scripts/dev_telemetry.py counts                 # counts per log level
"""
import argparse
import os
import sqlite3
import sys

DB_PATH = os.path.join(os.path.dirname(__file__), "..", "data", "universalops.db")


def connect():
    if not os.path.exists(DB_PATH):
        sys.exit(f"DB not found: {DB_PATH}")
    return sqlite3.connect(DB_PATH)


def dump_schema(conn):
    tables = [r[0] for r in conn.execute(
        "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")]
    print(f"TABLES ({len(tables)}): {', '.join(tables)}")
    for t in tables:
        cols = [r[1] for r in conn.execute(f"PRAGMA table_info('{t}')")]
        print(f"  {t}: {', '.join(cols)}")


def fetch(conn, sql, params=(), limit=25):
    cur = conn.execute(sql + " LIMIT ?", (*params, limit))
    rows = cur.fetchall()
    cols = [d[0] for d in cur.description]
    return cols, rows


def show(conn, cols, rows):
    if not rows:
        print("  (no rows)")
        return
    for r in rows:
        d = dict(zip(cols, r))
        ts = str(d.get("timestamp") or d.get("ts") or d.get("created_at") or "")[:23]
        level = str(d.get("level") or "")[:7].upper()
        msg = str(d.get("message") or d.get("msg") or d.get("detail") or "")
        print(f"{ts} | {level:7s} | {msg[:200]}")


def main():
    p = argparse.ArgumentParser()
    p.add_argument("mode", choices=["schema", "logs", "errors", "metrics", "counts"])
    p.add_argument("--limit", type=int, default=25)
    a = p.parse_args()
    conn = connect()
    if a.mode == "schema":
        dump_schema(conn)
        return
    if a.mode == "counts":
        for row in conn.execute("SELECT level, COUNT(*) FROM logs GROUP BY level ORDER BY 2 DESC"):
            print(f"  {row[0]:7s} {row[1]}")
        return
    if a.mode == "logs":
        cols, rows = fetch(conn, "SELECT * FROM logs ORDER BY id DESC", limit=a.limit)
    elif a.mode == "errors":
        cols, rows = fetch(conn,
            "SELECT * FROM logs WHERE level IN ('ERROR','WARN') ORDER BY id DESC", limit=a.limit)
    else:  # metrics
        try:
            cols, rows = fetch(conn,
                "SELECT * FROM metrics ORDER BY id DESC", limit=a.limit)
        except sqlite3.OperationalError:
            print("  (no metrics table)")
            return
    show(conn, cols, rows)


if __name__ == "__main__":
    main()
