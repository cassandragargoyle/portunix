# ISO 16355 - Quality Function Deployment (QFD)

## Overview

**ISO 16355** is an international standard that defines Quality Function Deployment (QFD) - a systematic method for translating customer requirements into appropriate technical requirements throughout product development.

QFD originated in Japan in the late 1960s and has become a fundamental tool for product planning and quality management worldwide.

## Standard Structure

ISO 16355 consists of multiple parts:

| Part | Title | Description |
|------|-------|-------------|
| ISO 16355-1 | General guidelines | Introduction and principles of QFD |
| ISO 16355-2 | VoC/VoS acquisition (non-quantitative) | Interviews, focus groups, observation |
| ISO 16355-3 | VoC/VoS acquisition (quantitative) | Surveys, statistical sampling, analytics |
| ISO 16355-4 | VoC/VoS analysis | Translation, prioritization, benchmarking |
| ISO 16355-5 | Solution strategy | Concept selection, technology deployment |
| ISO 16355-6 | Optimization | Parameter and tolerance design |
| ISO 16355-7 | Comprehensive QFD | Integration of all QFD activities |
| ISO 16355-8 | Commercialization | Market introduction and lifecycle |

## Key Concepts

### Voice of Customer (VoC) and Voice of Stakeholder (VoS)

The foundation of QFD is systematically capturing what customers and stakeholders want.

**Voice of Customer (VoC)** - direct customer/user requirements:
- **Stated needs** - explicitly expressed requirements
- **Implied needs** - expected but not stated
- **Latent needs** - unrecognized desires that delight customers

**Voice of Stakeholder (VoS)** - requirements from other stakeholders:
- **Regulatory** - compliance, legal requirements, standards
- **Business** - profitability, market positioning, strategy
- **Technical** - maintainability, scalability, security
- **Partners** - integration requirements, SLAs
- **Internal teams** - development, operations, support needs

### House of Quality (HoQ)

The primary tool of QFD is a matrix called "House of Quality":

```
┌─────────────────────────────────┐
│     Correlation Matrix          │  ← Technical correlations
│         (roof)                  │
├─────────┬───────────────────────┤
│         │  Technical            │
│ Customer│  Requirements         │  ← HOWs
│  Needs  │  (columns)            │
│ (rows)  ├───────────────────────┤
│         │  Relationship         │
│  WHATs  │  Matrix               │  ← Strength of relationships
│         │                       │
├─────────┼───────────────────────┤
│ Priority│  Technical Targets    │  ← Measurable specifications
└─────────┴───────────────────────┘
```

### Four Phases of QFD

1. **Product Planning** - VoC + VoS → Design requirements
2. **Part Deployment** - Design requirements → Part characteristics
3. **Process Planning** - Part characteristics → Process parameters
4. **Production Planning** - Process parameters → Production requirements

> **Note**: Traditional QFD (Akao, 1960s) focused primarily on customer needs. ISO 16355 extends the input to include both Voice of Customer (VoC) and Voice of Stakeholder (VoS) - regulatory, business, technical, and partner requirements.

## Application in Software Development

QFD principles apply to software products:

| QFD Concept | Software Application |
|-------------|---------------------|
| Customer needs | User stories, feature requests, feedback |
| Quality characteristics | Functional requirements, NFRs |
| Technical requirements | Architecture decisions, API design |
| Process parameters | Development practices, CI/CD |
| Production requirements | Deployment, monitoring, SLAs |

### MVP (Minimum Viable Product) and QFD

QFD supports iterative development through MVP approach:

1. **VoC/VoS Prioritization** - QFD helps identify which requirements are essential for MVP
2. **Must-have vs Nice-to-have** - House of Quality matrix reveals critical features
3. **Iterative Refinement** - Each MVP iteration feeds new VoC back into QFD process
4. **Risk Reduction** - Early validation of prioritized requirements

```
MVP Cycle with QFD:

  VoC/VoS → [QFD Prioritization] → MVP Features → Build → Release
                    ↑                                        │
                    └──────────── Feedback ←─────────────────┘
```

## Relevance to PTX-PFT

The `ptx-pft` helper implements QFD principles:

### Voice of Customer and Stakeholder Collection (VoC/VoS)
- Integration with feedback tools (Fider.io, Canny, ProductBoard)
- Systematic collection of user and stakeholder feedback
- Categorization and prioritization

### Requirements Traceability
- Bidirectional sync between feedback and documentation
- Linking customer requests to implementation
- Status tracking from request to delivery

### Continuous Improvement
- Feedback loop from users to developers
- Metrics and reporting on feedback handling
- Prioritization based on customer value

## Benefits of QFD

1. **Customer Focus** - Ensures products meet actual customer needs
2. **Reduced Development Time** - Fewer late-stage design changes
3. **Lower Costs** - Early identification of issues
4. **Better Communication** - Common framework for teams
5. **Documentation** - Clear record of design decisions
6. **Competitive Advantage** - Products that truly satisfy customers

## References

- ISO 16355-1:2015 - Application of statistical and related methods to new technology and product development process
- QFD Institute: https://www.qfdi.org/
- American Society for Quality (ASQ): https://asq.org/quality-resources/qfd-quality-function-deployment

## See Also

- [PTX-PFT Helper](../src/helpers/ptx-pft/README.md) - Implementation of QFD principles in Portunix
- [Issue #107](issues/internal/107-ptx-pft-product-feedback-tool-helper.md) - PTX-PFT implementation details
