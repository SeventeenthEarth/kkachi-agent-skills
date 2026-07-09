---
name: kkachi-design-review
description: Teal UI/UX design review gate for Kkachi workflow runs; applies only when project/task facts make Teal required.
version: 0.2.4
metadata:
  kas:
    role: teal
    suite_role: teal_design_reviewer
    required_gates:
      - DESIGN_PLAN_GATE
      - DESIGN_FIDELITY_REVIEW
---

# Kkachi Teal Design Review

Use this skill when a project or task declares a Teal lane and includes UI/UX scope. Teal is optional unless `project_has_teal_lane && ui_ux_change` is true.

## Role boundary

Teal reviews design/UX fitness, planned UI state, fidelity evidence, screenshots, and design acceptance references. Teal does not implement code, choose backend/runtime/provider/model settings, approve waivers alone, mutate KAH state, or replace Blue/Red/Orange/Gray/MAR gates.

For non-UI Kkachi source work, record `teal_required:false` and a concrete `teal_skip_reason`; do not inject Teal workflow nodes.

## Required evidence when Teal applies

When `teal_required:true`, require separate Teal evidence before the protected lifecycle points:

1. `DESIGN_PLAN_GATE` before implementation authorization.
2. `DESIGN_FIDELITY_REVIEW` before final acceptance.

Expected run evidence should preserve:

- `project_has_teal_lane:true`
- `ui_ux_change:true`
- derived `teal_required:true`
- Teal owner/profile or explicit missing-owner blocker
- design plan/spec reference
- fidelity criteria and screenshot/reference evidence
- Teal verdict and evidence refs
- waiver fields only when 주군-approved, bounded, scoped, and not expired

## No-substitution rule

Ordinary Red/Orange/Gray/Blue review, MAR, backend evidence, helper notes, and temporary subagents are not substitutes for required Teal verdict evidence. Blue commander full-suite possession of `kkachi-design-review` is operational context only; it does not make Blue, MAR, or temporary helper output an official Teal verdict when `teal_required=true`. If the Teal verdict is missing and no valid waiver exists, fail closed and report `required_teal_verdict_missing` or the KAH design-evidence reason code.

## KAH boundary

KAH validates deterministic `design-evidence.json` shape, evidence refs, skip/waiver fields, and fail-closed gate status. KAH does not judge design quality, select Teal owners, or install Teal skills. KAS owns Teal role installation and workflow policy.
