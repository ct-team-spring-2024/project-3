package internal

import "strings"

const (
	RED   = true
	BLACK = false
)

type Pair struct {
	Key   string
	Value []byte
}

type RBNode struct {
	Pair   Pair
	Color  bool
	Left   *RBNode
	Right  *RBNode
	Parent *RBNode
}

type RBTree struct {
	Root    *RBNode
	nilNode *RBNode
}

// NewRBTree creates a new Red-Black Tree with a sentinel nilNode
func NewRBTree() *RBTree {
	nilNode := &RBNode{Color: BLACK}
	return &RBTree{
		nilNode: nilNode,
		Root:    nilNode,
	}
}
func (tree *RBTree) Insert(pair Pair) {
	newNode := &RBNode{
		Pair:   pair,
		Color:  RED,
		Left:   tree.nilNode,
		Right:  tree.nilNode,
		Parent: tree.nilNode,
	}

	y := tree.nilNode
	x := tree.Root

	for x != tree.nilNode {
		y = x
		if strings.Compare(newNode.Pair.Key, x.Pair.Key) < 0 {
			x = x.Left
		} else {
			x = x.Right
		}
	}

	newNode.Parent = y
	if y == tree.nilNode {
		tree.Root = newNode
	} else if strings.Compare(newNode.Pair.Key, y.Pair.Key) < 0 {
		y.Left = newNode
	} else {
		y.Right = newNode
	}

	tree.fixInsert(newNode)
}
func (tree *RBTree) fixInsert(z *RBNode) {
	for z.Parent.Color == RED {
		if z.Parent == z.Parent.Parent.Left {
			y := z.Parent.Parent.Right
			if y.Color == RED {
				z.Parent.Color = BLACK
				y.Color = BLACK
				z.Parent.Parent.Color = RED
				z = z.Parent.Parent
			} else {
				if z == z.Parent.Right {
					z = z.Parent
					tree.leftRotate(z)
				}
				z.Parent.Color = BLACK
				z.Parent.Parent.Color = RED
				tree.rightRotate(z.Parent.Parent)
			}
		} else {
			y := z.Parent.Parent.Left
			if y.Color == RED {
				z.Parent.Color = BLACK
				y.Color = BLACK
				z.Parent.Parent.Color = RED
				z = z.Parent.Parent
			} else {
				if z == z.Parent.Left {
					z = z.Parent
					tree.rightRotate(z)
				}
				z.Parent.Color = BLACK
				z.Parent.Parent.Color = RED
				tree.leftRotate(z.Parent.Parent)
			}
		}
	}
	tree.Root.Color = BLACK
}

func (tree *RBTree) leftRotate(x *RBNode) {
	y := x.Right
	x.Right = y.Left
	if y.Left != tree.nilNode {
		y.Left.Parent = x
	}
	y.Parent = x.Parent
	if x.Parent == tree.nilNode {
		tree.Root = y
	} else if x == x.Parent.Left {
		x.Parent.Left = y
	} else {
		x.Parent.Right = y
	}
	y.Left = x
	x.Parent = y
}

func (tree *RBTree) rightRotate(x *RBNode) {
	y := x.Left
	x.Left = y.Right
	if y.Right != tree.nilNode {
		y.Right.Parent = x
	}
	y.Parent = x.Parent
	if x.Parent == tree.nilNode {
		tree.Root = y
	} else if x == x.Parent.Right {
		x.Parent.Right = y
	} else {
		x.Parent.Left = y
	}
	y.Right = x
	x.Parent = y
}
func (tree *RBTree) Delete(key string) {
	z := tree.searchNode(tree.Root, key)
	if z == tree.nilNode {
		return
	}

	y := z
	yOriginalColor := y.Color
	var x *RBNode

	if z.Left == tree.nilNode {
		x = z.Right
		tree.transplant(z, z.Right)
	} else if z.Right == tree.nilNode {
		x = z.Left
		tree.transplant(z, z.Left)
	} else {
		y = tree.minimum(z.Right)
		yOriginalColor = y.Color
		x = y.Right
		if y.Parent == z {
			x.Parent = y
		} else {
			tree.transplant(y, y.Right)
			y.Right = z.Right
			y.Right.Parent = y
		}
		tree.transplant(z, y)
		y.Left = z.Left
		y.Left.Parent = y
		y.Color = z.Color
	}

	if yOriginalColor == BLACK {
		tree.fixDelete(x)
	}
}
func (tree *RBTree) transplant(u, v *RBNode) {
	if u.Parent == tree.nilNode {
		tree.Root = v
	} else if u == u.Parent.Left {
		u.Parent.Left = v
	} else {
		u.Parent.Right = v
	}
	v.Parent = u.Parent
}

func (tree *RBTree) minimum(x *RBNode) *RBNode {
	for x.Left != tree.nilNode {
		x = x.Left
	}
	return x
}

func (tree *RBTree) searchNode(x *RBNode, key string) *RBNode {
	for x != tree.nilNode {
		cmp := strings.Compare(key, x.Pair.Key)
		if cmp == 0 {
			return x
		} else if cmp < 0 {
			x = x.Left
		} else {
			x = x.Right
		}
	}
	return tree.nilNode
}
func (tree *RBTree) fixDelete(x *RBNode) {
	for x != tree.Root && x.Color == BLACK {
		if x == x.Parent.Left {
			w := x.Parent.Right
			if w.Color == RED {
				w.Color = BLACK
				x.Parent.Color = RED
				tree.leftRotate(x.Parent)
				w = x.Parent.Right
			}
			if w.Left.Color == BLACK && w.Right.Color == BLACK {
				w.Color = RED
				x = x.Parent
			} else {
				if w.Right.Color == BLACK {
					w.Left.Color = BLACK
					w.Color = RED
					tree.rightRotate(w)
					w = x.Parent.Right
				}
				w.Color = x.Parent.Color
				x.Parent.Color = BLACK
				w.Right.Color = BLACK
				tree.leftRotate(x.Parent)
				x = tree.Root
			}
		} else {
			w := x.Parent.Left
			if w.Color == RED {
				w.Color = BLACK
				x.Parent.Color = RED
				tree.rightRotate(x.Parent)
				w = x.Parent.Left
			}
			if w.Right.Color == BLACK && w.Left.Color == BLACK {
				w.Color = RED
				x = x.Parent
			} else {
				if w.Left.Color == BLACK {
					w.Right.Color = BLACK
					w.Color = RED
					tree.leftRotate(w)
					w = x.Parent.Left
				}
				w.Color = x.Parent.Color
				x.Parent.Color = BLACK
				w.Left.Color = BLACK
				tree.rightRotate(x.Parent)
				x = tree.Root
			}
		}
	}
	x.Color = BLACK
}

// Update updates the value for a given key. Returns true if the key was found and updated.
func (tree *RBTree) Update(key string, newValue []byte) bool {
	node := tree.searchNode(tree.Root, key)
	if node == tree.nilNode {
		return false // Key not found
	}
	node.Pair.Value = newValue
	return true
}

func (tree *RBTree) ToSortedSlice() []Pair {
	var result []Pair
	tree.inOrder(tree.Root, &result)
	return result
}

func (tree *RBTree) inOrder(node *RBNode, result *[]Pair) {
	if node != tree.nilNode {
		tree.inOrder(node.Left, result)
		*result = append(*result, node.Pair)
		tree.inOrder(node.Right, result)
	}
}
