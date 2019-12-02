
SleepingState = "Sleeping"
FindState = "Find"
FoundState = "Found"

class Graph:
    def __init__(self):
        self.node_list = list()
        self.node_hash = dict()

    def add_node(self, id, state, ):
        self.node_list.append(Node(id, state))
        self.node_hash[id] = 1

    def has_node(self, node_id):
        return node_id in self.node_hash

class Node:
    def __init__(self, id, state):
        self.id = id
        self.state = state
        self.neighbors_list = list()

    def add_edge(self):
        pass

class Edge:
    def __init__(self, from_node, to_node, weight, state):
        self.from_node = from_node
        self.to_node = to_node
        self.weight = weight
        self.state = state


def make_graph():
    file = open('graph.txt', 'r')
    row_list = file.open().split('\n')
    for row in row_list:
        if row:
            if not row.startswith('graph') or not row.startswith('}'):
                begin, end = row.split(' -- ')
                node_id_1 = begin[-1]
                node_id_2 = end[0]
                print(node_id_1)
                print(node_id_2)

