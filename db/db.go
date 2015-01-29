package db

import (
	"errors"

	"github.com/ziahamza/blend"

	"github.com/nu7hatch/gouuid"
)

type Storage interface {
	Init(string) error

	Close()

	Drop() error

	// Fill in the details of the vertex by its Id.
	// The private details will only be available if the
	// private key is passed.
	GetVertex(*blend.Vertex) error

	/*
		QueryVertex(parent blend.Vertex, query string) (blend.Vertex, error)
	*/

	GetChildVertex(blend.Vertex, blend.Edge) (blend.Vertex, error)

	CreateChildVertex(v, vc *blend.Vertex, e blend.Edge) error

	// Adds a new vertex, only used to build root isolated vertices
	CreateVertex(*blend.Vertex) error

	// Updates the details of a vertex. An entire vertex needs to be given as all
	// details are updated at once. The update vertex automatically sets the private
	// key from the original vertex, it is not overwridden.
	UpdateVertex(*blend.Vertex) error

	DeleteVertex(*blend.Vertex) error

	DeleteVertexTree([]*blend.Vertex) error

	GetEdges(blend.Vertex, blend.Edge) ([]blend.Edge, error)

	// Add a specific edge to the DB. fills in the Edge pointer with the new ID
	// of the edge
	CreateEdge(blend.Vertex, *blend.Edge) error
}

var backend Storage

func Init(uri string, s Storage) error {
	backend = s
	return backend.Init(uri)
}

func Close() {
	backend.Close()
}

func Drop() error {
	return backend.Drop()
}

func GetEdges(v blend.Vertex, e blend.Edge) ([]blend.Edge, error) {
	return backend.GetEdges(v, e)
}

func GetVertex(vertex *blend.Vertex) error {
	if vertex.Id == "" {
		return errors.New("Vertex Id not passed")
	}

	return backend.GetVertex(vertex)
}

func GetChildVertex(v blend.Vertex, e blend.Edge) (blend.Vertex, error) {
	return backend.GetChildVertex(v, e)
}

func CreateChildVertex(v, vc *blend.Vertex, e blend.Edge) error {
	if vc.Id == "" {
		vid, err := uuid.NewV4()

		if err != nil {
			return err
		}

		vc.Id = string(vid.String())
	}
	return backend.CreateChildVertex(v, vc, e)
}

func CreateEdge(v blend.Vertex, edge *blend.Edge) error {
	if edge.Family != "ownership" && edge.Family != "private" &&
		edge.Family != "public" && edge.Family != "event" {
		return errors.New("Edge Family not supported")
	}

	// if no name given then make it unque by the edge_vertex
	// as edges are unique with respect to the name
	if edge.Name == "" {
		edge.Name = edge.To
	}

	return backend.CreateEdge(v, edge)
}

func CreateVertex(vertex *blend.Vertex) error {
	if vertex.Id == "" {
		vid, err := uuid.NewV4()

		if err != nil {
			return err
		}

		vertex.Id = string(vid.String())
	}

	err := backend.CreateVertex(vertex)

	if err == nil {
		err = PropogateChanges(*vertex, blend.Event{
			Source:  vertex.Id,
			Type:    "vertex:create",
			Created: vertex.LastChanged,
		})
	}

	return err
}

func UpdateVertex(vertex *blend.Vertex) error {
	if vertex.Id == "" {
		return errors.New("Vertex Id not passed")
	}

	return backend.UpdateVertex(vertex)
}

func DeleteVertex(vertex *blend.Vertex) error {
	return DeleteVertexTree([]*blend.Vertex{vertex})
}

func DeleteVertexTree(vertices []*blend.Vertex) error {
	return backend.DeleteVertexTree(vertices)
}

func PropogateChanges(vertex blend.Vertex, event blend.Event) error {
	return nil
}

func ConfirmVertex(vid string) bool {
	return backend.GetVertex(&blend.Vertex{Id: vid}) == nil
}

func ConfirmVertexKey(vid, vkey string) bool {
	vertex := &blend.Vertex{Id: vid, PrivateKey: vkey}
	err := backend.GetVertex(vertex)
	if err != nil {
		return false
	}

	return true
}
