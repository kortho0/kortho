package merkle

import (
	"bytes"
	"hash"
)

func GenHash(h hash.Hash, data []byte) []byte {
	h.Reset()
	h.Write(data)
	return h.Sum(nil)
}

// a new hash struct
func mkRoot(hash []byte) *MerkleRoot {
	return &MerkleRoot{
		root: hash,
	}
}

// create hash of leaf node
func mkLeafRootHash(h hash.Hash, data []byte) *MerkleRoot {
	return mkRoot(GenHash(h, data))
}

// create hash of branch node
func mkRootHash(h hash.Hash, a MKRoot, b MKRoot) *MerkleRoot {
	return mkRoot(GenHash(h, append(a.getMerkleRoot(), b.getMerkleRoot()...)))
}

func powerOfTwo0(n int) int {
	if n&(n-1) == 0 {
		return n
	} else {
		return powerOfTwo0(n & (n - 1))
	}
}

// calculate the minimum value of 2 ^ n for this party
func powerOfTwo(n int) int {
	if n&(n-1) == 0 {
		return n >> 1
	}
	return powerOfTwo0(n)
}

// create a new merkle tree
func New(hash hash.Hash, data [][]byte) *MerkleTree {
	var n int

	if data == nil || len(data) == 0 {
		return nil
	}
	if n = len(data); n == 0 {
		return nil
	}
	r := &MerkleTree{
		hash: hash,
	}
	r.tree = r.mkMerkleTreeRoot(n, data)
	return r
}

// creat proof path
func constructPath(proof MerkleProof, leaf MKNode, node MKNode) MerkleProof {
	if leaf.isLeafNode() {
		nodeRoot := node.getMKRoot()
		leaRoot := leaf.getMKRoot()
		if bytes.Compare(leaRoot.getMerkleRoot(), nodeRoot.getMerkleRoot()) == 0 {
			return proof
		}
		return MerkleProof{}
	}

	ln := leaf.leftNode()
	rn := leaf.rightNode()
	lProofElem := ProofElem{
		isLeft:      true,
		nodeRoot:    ln.getMKRoot(),
		siblingRoot: rn.getMKRoot(),
	}
	rProofElem := ProofElem{
		isLeft:      false,
		nodeRoot:    rn.getMKRoot(),
		siblingRoot: ln.getMKRoot(),
	}
	lpath := constructPath(append(proof, lProofElem), ln, node)
	rpath := constructPath(append(proof, rProofElem), rn, node)
	return append(lpath, rpath...)
}

// creat proof path
func (t *MerkleTree) merkleProof(node MKNode) MerkleProof {
	return constructPath(MerkleProof{}, t.GetMtRoot(), node)
}

// check whether the data is in a given proof path
func validate(hash hash.Hash, proof MerkleProof, root MKNode, leaf MKNode) bool {
	len := len(proof)
	if len == 0 {
		r := root.getMKRoot()
		leaRoot := leaf.getMKRoot()
		return bytes.Compare(leaRoot.getMerkleRoot(), r.getMerkleRoot()) == 0
	}

	if !(bytes.Compare(proof[len-1].nodeRoot.getMerkleRoot(), leaf.getMKRoot().getMerkleRoot()) == 0) {
		return false
	}

	var node *MerkleNode
	if proof[len-1].isLeft {
		node = &MerkleNode{
			root: mkRootHash(hash, proof[len-1].nodeRoot, proof[len-1].siblingRoot),
		}
	} else {
		node = &MerkleNode{
			root: mkRootHash(hash, proof[len-1].siblingRoot, proof[len-1].nodeRoot),
		}
	}
	return validate(hash, proof[:len-1], root, node)

}

// check whether the data is in a given proof path
func (t *MerkleTree) validateMerkleProof(proof MerkleProof, node MKNode) bool {
	return validate(t.hash, proof, t.GetMtRoot(), node)
}

// detects whether the data exists in the tree
func (t *MerkleTree) VerifyNode(data []byte) bool {
	node := t.mkLeaf(data)
	proof := t.merkleProof(node)
	return t.validateMerkleProof(proof, node)
}
