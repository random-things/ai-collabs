# AI/Human Collaborations

This repository is a collaborative research space where I work with various Large Language Models (LLMs) to explore novel algorithms and data structures.

## How It Works

1. **LLM Proposals**: LLMs suggest novel algorithms, data structures, or optimizations for specific problems
2. **LLM Review**: A reasoning LLM is prompted to review the original proposal
3. **Human Review**: I pick apart the ideas to the extent that I'm able, trying to find problems with the LLM's assumptions
4. **Collaborative Prototype**: We write a Python implementation as a smoke test to verify that a code implementation functions and has the properties that the LLM alleged. This stage focuses on:
   - Correctness verification
   - Edge case handling
   - Practical usability
5. **Collaborative Iteration**: When I find issues with the theoretical basis during implementation, I discuss it with multiple LLMs to try to find ways to fix or mitigate flaws
6. **Collaborative Implementation**: Finally, we write a Go implementation. This stage focuses on:
   - Performance benchmarking
7. (Rarely) **Collaborative Tuning**: If I'm convinced that substantial performance gains can be realized by moving some portion of the logic to Rust, we do that using CGO.

## Notes

### LLM.md

Each `LLM.md` file contains a note at the top about the status of that particular project.

* 🚫 = major flaw that I couldn't resolve or wasn't interested in resolving
* ⚠️ = works, but with caveats
* ✅ = works and lives up to performance guarantees
* ❔ = unexplored, work in progress

Most directories have an `LLM.md` file that contains the prompts and conversations used to create and discuss the algorithm/data structure. These are included purely for reference with respect to implementations contained in this repository. Do **not** trust proofs that LLMs have provided without independent verification. Similarly, my inability to make something work should not be taken as a guarantee that what the LLM proposed will not work.

Although I have not yet run across a case where what the LLM suggested worked, but was not as performant as advertised, there is always the possibility that performance issues are a problem with my implementation specifically.

### Completeness

I wasn't intending to publish my results originally, so I neglected to save every conversation. If pieces are missing, that is why. Additionally, I've gone through and edited LLM responses to some degree (usually to clean up formatting, fix LaTeX) so they are not necessarily verbatim what the LLM supplied.

### Failures

Failed ideas are included in this repository as a reference for how and why things can go wrong. It's interesting (at least to me) to see where LLMs have problems and to see their interpretation of their mistakes. Additionally, even if the ideas themselves aren't particularly sound, there's no telling what inspiration it might spark for someone reading it.

## Projects

| Name | Description | Viable | Meets Guarantees? | Python | Go | Rust |
| ---- | ----------- | ------ | ------ | --- | --- | --- |
| Constellation Search | Use two anchors for indexing exact string search | ✅ | ✅ | ✅ | ✅ | ✅ |
| VT Syndrome Prefilter | Use Varshamov-Tenengolts syndromes as keys for efficient fuzzy matching. | ✅ | ✅ | ✅ | ✅ | ❌ |
| Zeckendorf Skip List |  | ✅ | ✅ | ✅ | ✅ | ❌ |
| Gyre String Index | Use contiguous bit-runs to speed string search. | ❌ | ❌ | ✅ | ❌ | ❌ |
| Harmonic Ladder | Use harmonic decomposition to achieve constant-time insertion, retrieval, and deletion, with perfect order preservation and zero collisions or rebalancing.| ❌ | ❌ | ❌ | ❌ | ❌ |
| Harmonic Ladder / Quadratic-Residue Egyptian Decomposer | Egyptian fraction decomposition algorithm | ❌ | ❌ | ✅ | ❌ | ❌ |
| SparseHash Wheel |  | ❌ | ❌ | ✅ | ❌ | ❌ |

## Process

I start with a prompt:

>Think creatively and imaginatively to invent a typo-tolerant string comparison algorithm for search-as-you-type. It must be mathematically correct and should have a profound performance benefit over extant options. Consider analogous problems from other domains.

Originally, I made the mistake of going to a Python implementation immediately. That wasted a lot of time. Now, I use a reasoning LLM with the ability to execute code to evaluate the output:

>Test all of the lemmas and assumptions with Python to verify that they would work in practice, not just in theory.

This provides a decent enough first-pass and the LLM can, more often than not, find the flaws in its original proposal and assess whether the benefits originally promised still hold. If I'm able to create a mathematically correct prototype in Python, I move on to Go in order to test any performance claims. If the performance claims hold, I typically ask for an assessment of how useful the algorithm or data structure is in comparison to the ones currently used and whether there are other applications for the algorithm/data structure where it could improve on existing methods.

>How useful is the VT Prefix Filter in comparison to the current filters used for this application? Are there other situations that would benefit from its use?

## But why?

LLMs get looked down upon as stochastic parrots. While I'm not saying they can (or even should) replace humans in all endeavors, I do want to show that within the right framework, they can be capable of generating novel insights and, with human assistance (for now), develop those insights into solutions to real-world problems.

Many LLMs can propose interesting algorithmic solutions, but there's often a gap between:
- Theoretical proposals and practical implementations
- Expected performance and real-world behavior
- Algorithmic correctness and production readiness

This repository serves as a bridge, where I:
1. Implement LLM-proposed algorithms
2. Verify their theoretical properties
3. Benchmark their performance
4. Document their practical applications
5. Share the results

I see this project as an implementation of the idea that a million monkeys on a million typewriters will eventually produce the complete works of Shakespeare: if a lot of people are trying to solve problems with LLMs and go through the process of checking their work, some of the ideas are bound to be legitimate scientific advances and innovations.

## Contributing

While this is primarily a personal research space, I welcome:
- Discussion of the implemented algorithms/data structures
- Suggestions for new problems to explore
- Performance optimization ideas
- Real-world use cases

## License

This project is licensed under the MIT License - see the LICENSE file for details.
