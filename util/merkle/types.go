package merkle

import (
	"hash"
)

type MKRoot interface {
	getMerkleRoot() []byte // hashed
}

type MKNode interface {
	isLeafNode() bool
	leftNode() MKNode
	rightNode() MKNode
	getMKRoot() MKRoot
	getMKHash() []byte
}

type MKTree interface {
	GetMtHash() []byte // return mekle tree root hashed
	GetMtRoot() MKNode // return mekle tree root
	VerifyNode([]byte) bool
	mkLeaf(data []byte) MKNode          // create leaf
	mkBranch(left, right MKNode) MKNode // create  ranch
}

type MerkleRoot struct {
	root []byte
}

type MerkleNode struct {
	isLeaf bool
	data   []byte
	root   MKRoot
	left   *MerkleNode
	right  *MerkleNode
}

type MerkleTree struct {
	tree MKNode
	hash hash.Hash
}

type MerkleProof []ProofElem
type ProofElem struct {
	isLeft      bool
	nodeRoot    MKRoot
	siblingRoot MKRoot
}

// get hash
func (a *MerkleRoot) getMerkleRoot() []byte {
	return a.root
}

// left node
func (n *MerkleNode) leftNode() MKNode {
	return n.left
}

// right node
func (n *MerkleNode) rightNode() MKNode {
	return n.right
}

// get hash struct
func (n *MerkleNode) getMKRoot() MKRoot {
	return n.root
}

// get hash of node
func (n *MerkleNode) getMKHash() []byte {
	return n.getMKRoot().getMerkleRoot()
}

// is leaf
func (n *MerkleNode) isLeafNode() bool {
	return n.isLeaf
}

// get root of tree
func (t *MerkleTree) GetMtRoot() MKNode {
	if t == nil {
		return nil
	}
	return t.tree
}

// get root hash of tree
func (t *MerkleTree) GetMtHash() []byte {
	if t == nil {
		return nil
	}
	return t.GetMtRoot().getMKRoot().getMerkleRoot()
}

// create a leaf
func (t *MerkleTree) mkLeaf(data []byte) MKNode {
	if data == nil || len(data) == 0 {
		return nil
	}
	return &MerkleNode{
		isLeaf: true,
		data:   data,
		root:   mkLeafRootHash(t.hash, data),
	}
}

// create a branch
func (t *MerkleTree) mkBranch(l, r MKNode) MKNode {
	left, ok := l.(*MerkleNode)
	if !ok || left == nil {
		return nil
	}
	right, ok := r.(*MerkleNode)
	if !ok || right == nil {
		return nil
	}
	return &MerkleNode{
		isLeaf: false,
		left:   left,
		right:  right,
		root:   mkRootHash(t.hash, left.root, right.root),
	}
}

// create a new merkle tree by slice
func (t *MerkleTree) mkMerkleTreeRoot(n int, data [][]byte) MKNode {
	switch n {
	case 1:
		return t.mkLeaf(data[0])
	default:
		i := powerOfTwo(n)
		return t.mkBranch(t.mkMerkleTreeRoot(i, data[:i]), t.mkMerkleTreeRoot((n-i), data[i:]))
	}
}
