#!/bin/bash

# OS validation for local testing
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo "⚠️  WARNING: You are running on non-Linux OS but switching to Linux tester role"
    echo "   This role is designed for Linux testing and should only accept Linux tests on local host"
    echo "   For container/VM testing, accept tests based on container/VM OS, not host OS"
    echo ""
fi

ln -sfn tester-linux.md .claude/roles/current.md
echo "🧪  Switched to role: TESTER (Linux)"
echo ""
echo "📋 Role Guidelines:"
echo "   • Local host testing: Only accept Linux tests"
echo "   • Container/VM testing: Accept tests based on container/VM OS"
echo "   • Always document tested OS in acceptance protocol"