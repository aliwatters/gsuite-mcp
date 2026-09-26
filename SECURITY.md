# Security profile

<!-- deep-review-security-profile: hostile-public-service -->

This repository uses `hostile-public-service` as its review posture. The server exposes no public network endpoint; the label describes the trust boundaries around its data and capabilities, not its network topology.

Review tooling resolves a pull request's profile from its base commit. A pull request that introduces or changes this file is reviewed under the profile in its base; the new declaration applies to subsequent pull requests.

## In scope

- **Protected assets:** OAuth tokens for a real Google account across Gmail, Drive, and Calendar scopes, and the operator's mail, files, and calendar contents. Credential and token handling must protect those assets.
- **Untrusted content:** Anyone can send the operator an email or calendar invite, or share a Drive document. The content of those messages, invites, and documents is attacker-authored input.
- **Agent actions:** This MCP server exposes privileged capabilities to an agent whose context can contain that untrusted content. Instructions embedded in the content can influence calls to `send`, `draft`, or `download`; ordinary input validation alone does not address this prompt-injection boundary. Outward-facing or hard-to-reverse actions, including sending mail and changing calendar entries, warrant explicit review.
- **External code:** The repository is public and accepts external pull requests, so contributed code is within the review threat model.

The account-resolution path traversal reported in [#182](https://github.com/aliwatters/gsuite-mcp/issues/182), fixed by [#183](https://github.com/aliwatters/gsuite-mcp/pull/183) in v0.4.1, is a concrete example of a credential-boundary failure in scope.

## Out of scope

The local control plane retains a low-adversary assumption: the operator's own configuration, caches, filesystem, and processes running as that same user are trusted. A hostile local user is outside this profile. Symlink, no-follow, TOCTOU, and path-identity guards over those operator-owned files, or signatures and attestation between the operator's own processes, need a reproduced failure to justify their cost. PCI, cardholder-data, and payment controls are also outside this profile.

## Mitigation standard

Every proposed mitigation should identify a reproduced failure, the trust boundary it protects, and why a smaller control is insufficient. Describing a change only as "defensive" or "hardening" does not establish its scope.
