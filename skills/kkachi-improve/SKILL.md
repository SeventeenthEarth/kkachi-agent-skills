---
name: kkachi-improve
description: Capture Kkachi self-improvement candidates from real run evidence and decide whether they belong in run artifacts, project overlay, prompt profile, phase skill reference, script, or shared KHS.
version: 0.2.0
---

# Kkachi Improve

Use this skill at the end of a run or after a repeated process/prompt failure.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Improve from evidence, not speculation. KHS starts as a pre-template / seed
skill system; real Hermes/Kkachi runs mature it by routing observed lessons to
the smallest durable surface that fits. Project-local improvement comes before
shared KHS promotion unless the lesson is already general.

## Inputs

- run artifacts
- verification failures or retries
- backend prompt performance
- feedback triage
- `registries/improvement-promotion-policy.yaml`

## Outputs

- `improvement-note.md`
- promotion recommendation: `run-artifact|project-overlay|backend-prompt-profile|phase-skill-reference|script|shared-khs`
- skill QA requirement when shared KHS changes are proposed
