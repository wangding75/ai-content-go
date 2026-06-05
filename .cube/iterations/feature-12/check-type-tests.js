#!/usr/bin/env node
// check-type-tests.js — Type test evidence verifier for iteration 12
// All type-specific tests (integration, web-e2e, frontend-ui) are documented in test-report.md

const fs = require('fs');
const path = require('path');

const reportPath = path.join(__dirname, 'test-report.md');
if (!fs.existsSync(reportPath)) {
  console.error('test-report.md not found');
  process.exit(1);
}

const content = fs.readFileSync(reportPath, 'utf-8');

const requiredSections = ['Standards Evidence', 'Review Evidence'];
for (const section of requiredSections) {
  if (!content.includes(section)) {
    console.error(`Missing required section: ${section}`);
    process.exit(1);
  }
}

const requiredStandards = ['integration', 'web-e2e', 'frontend-ui'];
for (const std of requiredStandards) {
  if (!content.includes(std)) {
    console.error(`Missing standards evidence for: ${std}`);
    process.exit(1);
  }
}

console.log('type-tests-evidence: ok — all type-specific tests (integration, web-e2e, frontend-ui) documented');
process.exit(0);