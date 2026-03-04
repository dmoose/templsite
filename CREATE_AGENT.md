# How to Create an AGENT.md File for Any Project

This guide provides detailed instructions for AI agents (LLMs) to create comprehensive `AGENT.md` documentation files that enable context continuity across sessions.

## Purpose

An `AGENT.md` file serves as a complete knowledge base that allows any AI agent to:
- Understand a project's architecture and design decisions
- Resume work without prior context
- Make consistent decisions aligned with project philosophy
- Navigate the codebase effectively
- Follow established patterns and conventions

## When to Create AGENT.md

Create this file when:
- Starting a new project with staged development
- Reaching a milestone or completion of major features
- Before pausing development for an extended period
- After implementing complex architecture that needs documentation
- When onboarding requirements become clear

## Document Structure

### 1. Header and Purpose
```markdown
# Agent Context Document

This document provides complete context for AI agents working on [PROJECT_NAME].
```

**Include:**
- Document purpose statement
- Target audience (AI agents)
- How to use this document

### 2. Project Overview

**Essential Elements:**
- Project name and one-line description
- Core purpose and problem it solves
- Key philosophy or design principles (3-5 bullet points)
- Technology choices and why they were made

**Example:**
```markdown
## Project Overview

**ProjectName** is a [type of application] built with [main technology]. 
It [primary purpose] while [key differentiator].

### Core Philosophy
- **Principle 1**: Brief explanation
- **Principle 2**: Brief explanation
- **Principle 3**: Brief explanation
```

### 3. Project Status

**Critical Information:**
- Current version or milestone
- Completed features/stages (with checkmarks ✅)
- In-progress work (with ⏳)
- Next planned work (with ⏭️)
- Reference to detailed roadmap if exists

**Example:**
```markdown
### Current Version: v0.3.0

### Completed
- ✅ Feature A (v0.1.0)
- ✅ Feature B (v0.2.0)
- ✅ Feature C (v0.3.0)

### Next
- ⏭️ Feature D (Planned)
```

### 4. Project Structure

**Provide:**
- Complete directory tree with ASCII art
- Purpose of each major directory
- Key files and what they contain
- Naming conventions used

**Format:**
```markdown
## Project Structure

\`\`\`
project/
├── dir1/              # Purpose of dir1
│   ├── subdir/       # Purpose of subdir
│   │   └── file.ext  # What this file does
│   └── file.ext      # What this file does
├── dir2/              # Purpose of dir2
└── file.ext           # What this file does
\`\`\`
```

### 5. Key Technologies

**Document:**
- Primary dependencies with versions
- External tools or CLIs required
- Development tools needed
- Why each technology was chosen (briefly)

**Example:**
```markdown
## Key Technologies

### Dependencies
\`\`\`
library/package v1.2.3    // Purpose and why chosen
another/lib v2.0.0        // Purpose and why chosen
\`\`\`

### External Tools
- **Tool Name**: Usage and installation method
```

### 6. Architecture & Core Concepts

**Include for each major system/component:**

#### 6.1 Data Structures
```markdown
### [Component Name]

\`\`\`[language]
type/struct Definition {
    field: type  // Purpose
    field: type  // Purpose
}
\`\`\`

**Purpose**: What this structure represents
**Key Methods**: Brief list with purposes
```

#### 6.2 Key Algorithms/Processes
- Workflow diagrams in text
- Step-by-step processes
- Decision trees
- State machines

#### 6.3 Design Patterns
- Patterns used and where
- Why they were chosen
- How they're implemented

### 7. Configuration System

**Document:**
- Configuration file format
- All configuration options with types
- Default values (critical!)
- Validation rules
- Example configuration

**Example:**
```markdown
## Configuration

### Structure
\`\`\`[language]
// Show actual config structure
\`\`\`

### Default Values
- Option1: "default" - Purpose
- Option2: 123 - Purpose
```

### 8. Core Workflows

**For each major operation:**
```markdown
### [Operation Name]

#### Input
- What it receives
- Expected format
- Validation performed

#### Process
1. Step 1 with details
2. Step 2 with details
3. Step 3 with details

#### Output
- What it produces
- Where it goes
- Format/structure

#### Error Handling
- Common errors
- How they're handled
```

### 9. Development Guide

**Include:**

#### 9.1 Common Tasks
```markdown
### Adding a New Feature
1. Step 1
2. Step 2
3. Step 3

### Fixing a Bug
1. Step 1
2. Step 2
```

#### 9.2 Testing Strategy
- Testing philosophy
- Coverage targets
- Test patterns used
- How to run tests
- Example test structure

#### 9.3 Code Conventions
- Naming conventions
- File organization
- Comment styles
- Import ordering
- Any project-specific patterns

### 10. Build System

**Document:**
- Build commands and what they do
- Development workflow
- Deployment process
- Environment variables
- Dependencies management

**Example:**
```markdown
## Build System

### Commands
- `make build` - What it does
- `make test` - What it does

### Workflow
1. Development step
2. Testing step
3. Deployment step
```

### 11. Git Workflow

**Essential:**
- Branch naming conventions
- Commit message format
- PR/merge process
- Tagging strategy
- When to create branches

**Example:**
```markdown
## Git Workflow

### Branch Strategy
\`\`\`bash
git checkout -b type/description
# Work
git commit -m "Format: Description"
git merge into main
git tag vX.Y.Z
\`\`\`
```

### 12. Important Notes

**Include:**
- Design decisions and rationale
- Trade-offs made
- Known limitations
- Technical debt
- Future considerations
- Things NOT to do and why

### 13. Next Steps / Future Work

**Provide:**
- Immediate next tasks
- Detailed breakdown of next stage
- Key considerations
- Potential challenges
- Resources needed

### 14. Resources

**Link to:**
- Internal documentation
- External API docs
- Tutorials or guides
- Design documents
- Related projects

### 15. Quick Reference

**Provide:**
- Common command cheatsheet
- Debugging tips
- FAQ
- Troubleshooting guide
- Contact information (if applicable)

## Best Practices for Writing AGENT.md

### DO:
✅ **Be Exhaustive**: Include everything needed to understand the project  
✅ **Use Examples**: Show actual code, commands, and structures  
✅ **Explain Why**: Don't just say what, explain why decisions were made  
✅ **Keep Current**: Update after each major milestone  
✅ **Use Formatting**: Headers, lists, code blocks, tables for clarity  
✅ **Link Liberally**: Reference other docs when they exist  
✅ **Include Context**: Explain the current state and how you got there  
✅ **Show Patterns**: Demonstrate coding patterns with examples  
✅ **Document Gotchas**: Note tricky areas and common mistakes  

### DON'T:
❌ **Assume Knowledge**: Explain even "obvious" things  
❌ **Be Vague**: Avoid "etc." without examples  
❌ **Skip Basics**: Include project setup and basics  
❌ **Ignore History**: Context of past decisions matters  
❌ **Use Relative Terms**: Say "current version is X" not "recent version"  
❌ **Omit Dependencies**: List ALL dependencies with versions  
❌ **Forget Edge Cases**: Document error handling and special cases  

## Section-by-Section Guidelines

### Project Overview
- **Length**: 2-4 paragraphs
- **Focus**: What, why, and how (high level)
- **Include**: Core technologies and approach
- **Avoid**: Implementation details (save for later sections)

### Project Status
- **Length**: 1 page max
- **Focus**: Current state, clear progress indicators
- **Include**: Version numbers, completion percentages if applicable
- **Update**: After every milestone

### Project Structure
- **Length**: 1-2 pages
- **Focus**: Physical layout of code
- **Include**: Every major directory, key files
- **Format**: Use ASCII tree, add comments

### Architecture
- **Length**: 2-4 pages per major component
- **Focus**: How things work together
- **Include**: Data flow, interfaces, patterns
- **Use**: Diagrams (even ASCII art)

### Workflows
- **Length**: 1 page per workflow
- **Focus**: Step-by-step processes
- **Include**: Inputs, outputs, error handling
- **Format**: Numbered lists, code examples

### Development Guide
- **Length**: 2-3 pages
- **Focus**: How to actually work on the project
- **Include**: Common tasks, gotchas, patterns
- **Practical**: Real examples, not theory

## Maintenance

### When to Update
- After completing any stage/milestone
- When architecture changes
- When dependencies are added/updated
- When workflows change
- When bugs reveal misunderstandings

### What to Update
- Status section (always)
- Completed features list
- Next steps section
- Version numbers
- Any changed workflows
- New patterns or conventions

### Version the Document
Consider adding to the bottom:
```markdown
---
**Last Updated**: YYYY-MM-DD
**Current Version**: vX.Y.Z
**Next Review**: After Stage N completion
```

## Template Checklist

Use this checklist when creating AGENT.md:

```markdown
- [ ] Project overview with core philosophy
- [ ] Current status with completed/in-progress/next items
- [ ] Complete project structure with annotations
- [ ] Key technologies with versions and purposes
- [ ] All major data structures documented
- [ ] Core workflows explained step-by-step
- [ ] Configuration system fully documented
- [ ] Build system commands and usage
- [ ] Git workflow and conventions
- [ ] Testing strategy and examples
- [ ] Common development tasks
- [ ] Design decisions and rationale
- [ ] Known limitations
- [ ] Next steps with details
- [ ] Resource links
- [ ] Quick reference commands
- [ ] Debugging tips
- [ ] Last updated date
```

## Example Outline

Here's a complete outline you can follow:

```markdown
# Agent Context Document

## Project Overview
[2-3 paragraphs + core philosophy]

## Project Status
[Current version, completed features, next steps]

## Project Structure
[Directory tree with annotations]

## Key Technologies
[Dependencies and tools with versions]

## [Component 1] Architecture
[Data structures, workflows, patterns]

## [Component 2] Architecture
[Data structures, workflows, patterns]

## Configuration System
[Structure, defaults, validation]

## Build Workflow
[Commands, processes, deployment]

## Development Guide
[Common tasks, testing, conventions]

## Git Workflow
[Branching, commits, tagging]

## Important Notes
[Decisions, trade-offs, limitations]

## Next Steps
[Detailed next stage plan]

## Resources
[Links to docs and guides]

## Quick Reference
[Commands, debugging, FAQ]

---
**Last Updated**: YYYY-MM-DD
```

## Tips for Success

1. **Start Early**: Create AGENT.md as soon as architecture is clear
2. **Update Often**: After each significant commit or milestone
3. **Be Specific**: Use actual code, not pseudocode
4. **Think Fresh**: Write for someone with zero context
5. **Test It**: Can you resume work from just this doc?
6. **Include Failures**: Document what didn't work and why
7. **Show Evolution**: Explain how current design was reached
8. **Link Everything**: Reference other docs liberally
9. **Use Tools**: Leverage diagrams, tables, code blocks
10. **Get Feedback**: If working with others, validate it's clear

## Common Mistakes to Avoid

1. **Too Brief**: "See code for details" defeats the purpose
2. **Too Technical**: Balance detail with readability
3. **Outdated**: An old AGENT.md is worse than none
4. **Missing Context**: Explain *why* not just *what*
5. **No Examples**: Abstract descriptions aren't enough
6. **Scattered Info**: Keep related info together
7. **Unclear Next Steps**: Be specific about what's next
8. **No Status**: Always know what's done and what's not

## Validation

Before considering AGENT.md complete, verify:

- [ ] A new AI agent could understand the project in 15 minutes
- [ ] All major systems are documented
- [ ] Current status is crystal clear
- [ ] Next steps are actionable
- [ ] Code examples compile/run
- [ ] Commands are copy-pasteable
- [ ] All design decisions are explained
- [ ] No dangling references to missing docs
- [ ] Document is well-formatted and readable
- [ ] All technical terms are explained

## Final Note

A great AGENT.md file is:
- **Comprehensive** but not overwhelming
- **Technical** but readable
- **Current** and maintained
- **Practical** with real examples
- **Context-rich** with decision rationale
- **Future-focused** with clear next steps

The goal is to eliminate the "context-loading" time when resuming work or onboarding new contributors (human or AI). Every hour spent writing good AGENT.md documentation saves many hours of confusion later.

---

**This document is meta**: It's a guide for creating AGENT.md files. The AGENT.md file itself should be project-specific and follow the structure outlined here.