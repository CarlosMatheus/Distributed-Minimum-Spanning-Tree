from graphviz import Digraph

SleepingState = "Sleeping"
FindState = "Find"
FoundState = "Found"

RejectedState = "Rejected"
BranchState = "Branch"
BasicState = "Basic"

state_to_color = {
    SleepingState: "aquamarine",
    FindState: "deepskyblue1",
    FoundState: "gold1"
}


class Graph:
    def __init__(self):
        self.node_list = list()
        self.node_hash = dict()
        self.edge_list = list()
        self.edge_hash = dict()

    def add_edge(self, edge):
        self.edge_list.append(edge)
        self.edge_hash[edge.weight] = edge

    def add_node(self, id, state):
        node = Node(id, state, self)
        self.node_list.append(node)
        self.node_hash[id] = node
        return node

    def has_node_id(self, node_id):
        return node_id in self.node_hash

    def get_node_by_id(self, node_id):
        return self.node_hash[node_id]

    def print_graph(self):
        f = Digraph('graph', filename='graph.gv')

        for node in self.node_list:
            # f.attr(style='filled', color=state_to_color[node.state])
            f.node(node.id, style='filled', color=state_to_color[node.state])

        for edge in self.edge_list:
            f.edge(edge.from_node.id, edge.to_node.id, label=edge.weight)

        f.view()


class Node:
    def __init__(self, id, state, graph):
        self.id = id
        self.state = state
        self.neighbors_list = list()
        self.graph = graph

    def add_edge(self, edge):
        self.neighbors_list.append(edge)
        self.graph.add_edge(edge)


class Edge:
    def __init__(self, from_node, to_node, weight, state):
        self.from_node = from_node
        self.to_node = to_node
        self.weight = weight
        self.state = state


def make_graph():

    graph = Graph()

    file = open('graph.txt', 'r')
    row_list = file.read().split('\n')
    for row in row_list:
        if row:
            if not row.startswith('graph') and not row.startswith('}'):
                begin, end = row.split(' -- ')
                node_id_1 = begin[-1]
                node_id_2 = end[0]
                print(node_id_1)
                print(node_id_2)

                if not graph.has_node_id(node_id_1):
                    node_1 = graph.add_node(node_id_1, SleepingState)
                else:
                    node_1 = graph.get_node_by_id(node_id_1)

                if not graph.has_node_id(node_id_2):
                    node_2 = graph.add_node(node_id_2, SleepingState)
                else:
                    node_2 = graph.get_node_by_id(node_id_2)

                _, label, _, weight, _ = end.split('"')
                print(label)
                print(weight)

                edge1 = Edge(node_1, node_2, weight, BasicState)
                edge2 = Edge(node_2, node_1, weight, BasicState)

                node_1.add_edge(edge1)
                node_2.add_edge(edge2)

    return graph


graph = make_graph()
graph.print_graph()


