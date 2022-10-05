The preferred approach to implementing the AMD64 architecture, by a team eager for quickly arriving at a working model that could be displayed on a screen, was to establish **setOpcode**, **setModRM** and **setSIBByte** methods, and go filling it with "if" statements until there was a branch for every possible instruction byte. roughly 30% of the way into implementing the primary opcode map, the three methods became very difficult to read.

From this commit, we propose a different, more fundamental, mapped approach.

The opcode map will be implemented as a a map[byte] of 256 recursive **InstructionMapByte** structs.

The InstructionMapByte has the following fields:

 - isPrefix bool: whether the byte represents a prefix
 - mnemonic string: if the byte can define a mnemonic, this should be nonempty