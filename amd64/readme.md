The preferred approach to implementing the AMD64 architecture, by a team eager for quickly arriving at a working model that could be displayed on a screen, was writing three functions, namely **setOpcode**, **setModRM** and **setSIBByte** methods, and go filling it with "if" statements until there was a branch for every possible instruction byte. roughly 30% of the way into implementing the primary opcode map, the three methods became very difficult to read.

From this commit, we propose a different, more fundamental, mapped approach.

The opcode map will be implemented as a a map[byte] of 256 **opcode** structs.