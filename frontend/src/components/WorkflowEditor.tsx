import { useState, useCallback, useEffect, useRef } from 'react';
import {
  ReactFlow, ReactFlowProvider, addEdge, useNodesState, useEdgesState,
  Controls, Background, BackgroundVariant, Handle, Position,
  type Node, type Edge, type Connection, type NodeProps, useReactFlow,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Button, Space, Input, InputNumber, Select, Typography, Tooltip, Message } from '@arco-design/web-react';
import { IconSave, IconClose, IconDelete, IconDragDot } from '@arco-design/web-react/icon';

// ─── Backend DAG types ────────────────────────────────────────────────────────
type WFNodeType = 'acme_issue' | 'deploy_cert' | 'dns_validate' | 'notify' | 'condition' | 'wait';
interface WFNodeDef {
  id: string;
  type: WFNodeType;
  params?: Record<string, unknown>;
  depends_on?: string[];
}

// ─── Node palette config ──────────────────────────────────────────────────────
const NODE_META: Record<WFNodeType, {
  label: string; color: string; bg: string; icon: string;
  fields: { key: string; label: string; type: 'text' | 'number' | 'select' | 'textarea'; options?: string[]; placeholder?: string }[];
}> = {
  acme_issue: {
    label: '申请证书', color: '#165DFF', bg: '#EAF0FF', icon: '🔐',
    fields: [
      { key: 'domain_name', label: '域名', type: 'text', placeholder: 'example.com' },
      { key: 'acme_provider', label: 'ACME 提供商', type: 'select', options: ['letsencrypt', 'zerossl'] },
      { key: 'challenge', label: '验证方式', type: 'select', options: ['dns-01'] },
      { key: 'dns_account_id', label: 'DNS 账户 ID', type: 'number' },
      { key: 'account_email', label: '联系邮箱', type: 'text', placeholder: 'admin@example.com' },
      { key: 'staging', label: '测试环境 (staging)', type: 'select', options: ['false', 'true'] },
    ],
  },
  deploy_cert: {
    label: '部署证书', color: '#7816FF', bg: '#F4E8FF', icon: '🚀',
    fields: [
      { key: 'cert_source_node', label: '来源节点 ID（留空取上游）', type: 'text', placeholder: '上游 acme_issue 节点 ID' },
      { key: 'target_node_id', label: '目标 CDN 节点 ID', type: 'number' },
    ],
  },
  dns_validate: {
    label: '检查 DNS', color: '#FF7D00', bg: '#FFF3E8', icon: '🌐',
    fields: [
      { key: 'domain_name', label: '域名', type: 'text', placeholder: 'example.com' },
      { key: 'record_type', label: '记录类型', type: 'select', options: ['A', 'AAAA', 'CNAME', 'TXT', 'MX'] },
      { key: 'expected_value', label: '期望值', type: 'text', placeholder: 'IP 地址或目标值' },
      { key: 'timeout_seconds', label: '超时（秒）', type: 'number' },
    ],
  },
  notify: {
    label: '发送通知', color: '#00B42A', bg: '#E8FFE8', icon: '🔔',
    fields: [
      { key: 'channel_id', label: '通知渠道 ID', type: 'number' },
      { key: 'message', label: '消息内容', type: 'textarea', placeholder: '工作流已完成' },
    ],
  },
  condition: {
    label: '条件判断', color: '#F7BA1E', bg: '#FFFAE8', icon: '🔀',
    fields: [
      { key: 'expression', label: '条件表达式', type: 'text', placeholder: 'input.status == "success"' },
    ],
  },
  wait: {
    label: '等待', color: '#86909C', bg: '#F2F3F5', icon: '⏱',
    fields: [
      { key: 'seconds', label: '等待时长（秒）', type: 'number' },
    ],
  },
};

// ─── Custom node component ────────────────────────────────────────────────────
function WFNode({ id, data, selected }: NodeProps) {
  const meta = NODE_META[data.nodeType as WFNodeType];
  if (!meta) return null;
  return (
    <div style={{
      background: selected ? meta.bg : '#fff',
      border: `2px solid ${selected ? meta.color : '#E5E6EB'}`,
      borderRadius: 10,
      minWidth: 160,
      cursor: 'default',
      boxShadow: selected ? `0 0 0 3px ${meta.color}33` : '0 2px 8px #00000014',
      transition: 'box-shadow .15s, border-color .15s',
    }}>
      <Handle type="target" position={Position.Top}
        style={{ background: '#C9CDD4', border: '2px solid #fff', width: 10, height: 10, top: -6 }} />
      <div style={{ padding: '10px 14px', display: 'flex', alignItems: 'center', gap: 8 }}>
        <span style={{ fontSize: 20, lineHeight: 1 }}>{meta.icon}</span>
        <div style={{ minWidth: 0 }}>
          <div style={{ fontWeight: 700, fontSize: 13, color: meta.color, lineHeight: 1.3 }}>{meta.label}</div>
          <div style={{ fontSize: 11, color: '#86909C', marginTop: 2, lineHeight: 1.3, maxWidth: 110, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {(data.params as Record<string, unknown>)?.domain_name as string
              || (data.params as Record<string, unknown>)?.message as string
              || id}
          </div>
        </div>
      </div>
      <Handle type="source" position={Position.Bottom}
        style={{ background: meta.color, border: '2px solid #fff', width: 10, height: 10, bottom: -6 }} />
    </div>
  );
}

const NODE_TYPES = { workflowNode: WFNode };

// ─── Layout helpers ───────────────────────────────────────────────────────────
function autoLayout(defs: WFNodeDef[]): Node[] {
  // compute depth per node (max depth of dependencies)
  const depthMap: Record<string, number> = {};
  const compute = (id: string, visited = new Set<string>()): number => {
    if (id in depthMap) return depthMap[id];
    if (visited.has(id)) return 0;
    visited.add(id);
    const def = defs.find(d => d.id === id);
    const deps = def?.depends_on ?? [];
    const depth = deps.length === 0 ? 0 : Math.max(...deps.map(d => compute(d, visited) + 1));
    depthMap[id] = depth;
    return depth;
  };
  defs.forEach(d => compute(d.id));

  // group by depth, assign x/y
  const cols: Record<number, string[]> = {};
  Object.entries(depthMap).forEach(([id, d]) => {
    cols[d] = cols[d] ?? [];
    cols[d].push(id);
  });

  return defs.map(def => {
    const depth = depthMap[def.id] ?? 0;
    const col = cols[depth] ?? [];
    const row = col.indexOf(def.id);
    return {
      id: def.id,
      type: 'workflowNode',
      position: { x: depth * 240 + 60, y: row * 140 + 60 },
      data: { nodeType: def.type, params: def.params ?? {} },
    };
  });
}

function graphJSONToFlow(json: string): { nodes: Node[]; edges: Edge[] } {
  try {
    const defs: WFNodeDef[] = JSON.parse(json);
    if (!Array.isArray(defs)) throw new Error('not array');
    const nodes = autoLayout(defs);
    const edges: Edge[] = [];
    defs.forEach(def => {
      (def.depends_on ?? []).forEach(depId => {
        edges.push({
          id: `e-${depId}-${def.id}`,
          source: depId,
          target: def.id,
          animated: true,
          style: { stroke: '#165DFF', strokeWidth: 2 },
        });
      });
    });
    return { nodes, edges };
  } catch {
    return { nodes: [], edges: [] };
  }
}

function flowToGraphJSON(nodes: Node[], edges: Edge[]): string {
  const defs: WFNodeDef[] = nodes.map(n => ({
    id: n.id,
    type: n.data.nodeType as WFNodeType,
    params: (n.data.params as Record<string, unknown>) ?? {},
    depends_on: edges.filter(e => e.target === n.id).map(e => e.source),
  }));
  return JSON.stringify(defs);
}

// ─── Params panel (right side) ────────────────────────────────────────────────
function ParamsPanel({
  node, onChange, onDelete,
}: {
  node: Node;
  onChange: (id: string, params: Record<string, unknown>) => void;
  onDelete: (id: string) => void;
}) {
  const meta = NODE_META[node.data.nodeType as WFNodeType];
  const params = (node.data.params ?? {}) as Record<string, unknown>;

  const set = (key: string, val: unknown) => onChange(node.id, { ...params, [key]: val });

  if (!meta) return null;
  return (
    <div style={{ padding: '16px 12px', display: 'flex', flexDirection: 'column', gap: 10, height: '100%', overflow: 'auto' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
        <span style={{ fontSize: 20 }}>{meta.icon}</span>
        <Typography.Text bold style={{ color: meta.color, fontSize: 14 }}>{meta.label}</Typography.Text>
        <div style={{ flex: 1 }} />
        <Tooltip content="删除节点">
          <Button size="mini" type="text" status="danger" icon={<IconDelete />} onClick={() => onDelete(node.id)} />
        </Tooltip>
      </div>
      <div style={{ fontSize: 11, color: '#86909C', background: '#F7F8FA', borderRadius: 6, padding: '4px 8px' }}>
        节点 ID：<code style={{ fontSize: 11 }}>{node.id}</code>
      </div>
      {meta.fields.map(f => (
        <div key={f.key}>
          <div style={{ fontSize: 12, color: '#4E5969', marginBottom: 4, fontWeight: 500 }}>{f.label}</div>
          {f.type === 'select' ? (
            <Select
              size="small"
              value={params[f.key] as string ?? ''}
              onChange={v => set(f.key, v)}
              style={{ width: '100%' }}
            >
              {(f.options ?? []).map(o => <Select.Option key={o} value={o}>{o}</Select.Option>)}
            </Select>
          ) : f.type === 'number' ? (
            <InputNumber
              size="small"
              value={params[f.key] as number ?? undefined}
              onChange={v => set(f.key, v)}
              style={{ width: '100%' }}
              placeholder="0"
            />
          ) : f.type === 'textarea' ? (
            <Input.TextArea
              rows={2}
              value={params[f.key] as string ?? ''}
              onChange={v => set(f.key, v)}
              placeholder={f.placeholder}
              style={{ fontSize: 12 }}
            />
          ) : (
            <Input
              size="small"
              value={params[f.key] as string ?? ''}
              onChange={v => set(f.key, v)}
              placeholder={f.placeholder}
            />
          )}
        </div>
      ))}
    </div>
  );
}

// ─── Inner editor (needs ReactFlowProvider context) ───────────────────────────
let nodeCounter = 1;
function newId() { return `n${Date.now()}_${nodeCounter++}`; }

function EditorInner({
  initialNodes, initialEdges, onSave, onClose,
  workflowName, onNameChange,
}: {
  initialNodes: Node[];
  initialEdges: Edge[];
  onSave: (graphJSON: string) => void;
  onClose: () => void;
  workflowName: string;
  onNameChange: (v: string) => void;
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const { screenToFlowPosition } = useReactFlow();
  const canvasRef = useRef<HTMLDivElement>(null);

  const selectedNode = nodes.find(n => n.id === selectedNodeId) ?? null;

  const onConnect = useCallback(
    (conn: Connection) => setEdges(eds => addEdge({
      ...conn, animated: true,
      style: { stroke: '#165DFF', strokeWidth: 2 },
    }, eds)),
    [setEdges],
  );

  const addNode = (type: WFNodeType) => {
    const rect = canvasRef.current?.getBoundingClientRect();
    const cx = rect ? rect.width / 2 : 300;
    const cy = rect ? rect.height / 2 : 200;
    const pos = screenToFlowPosition({ x: (rect?.left ?? 0) + cx, y: (rect?.top ?? 0) + cy });
    const id = newId();
    const node: Node = {
      id, type: 'workflowNode',
      position: { x: pos.x + (Math.random() - 0.5) * 60, y: pos.y + (Math.random() - 0.5) * 60 },
      data: { nodeType: type, params: {} },
    };
    setNodes(nds => [...nds, node]);
    setSelectedNodeId(id);
  };

  const updateNodeParams = (id: string, params: Record<string, unknown>) => {
    setNodes(nds => nds.map(n => n.id === id ? { ...n, data: { ...n.data, params } } : n));
  };

  const deleteNode = (id: string) => {
    setNodes(nds => nds.filter(n => n.id !== id));
    setEdges(eds => eds.filter(e => e.source !== id && e.target !== id));
    setSelectedNodeId(null);
  };

  const handleSave = () => {
    if (nodes.length === 0) { Message.warning('至少需要一个节点'); return; }
    onSave(flowToGraphJSON(nodes, edges));
  };

  // Delete key handler
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.key === 'Delete' || e.key === 'Backspace') && selectedNodeId
          && !(e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) {
        deleteNode(selectedNodeId);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [selectedNodeId]); // eslint-disable-line

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', background: '#fff' }}>
      {/* ── Top toolbar ── */}
      <div style={{
        display: 'flex', alignItems: 'center', gap: 12, padding: '10px 16px',
        borderBottom: '1px solid #E5E6EB', background: '#fff', flexShrink: 0,
      }}>
        <span style={{ fontWeight: 700, fontSize: 15, color: '#1D2129', whiteSpace: 'nowrap' }}>工作流编辑器</span>
        <Input
          value={workflowName}
          onChange={onNameChange}
          placeholder="工作流名称（必填）"
          style={{ width: 240 }}
        />
        <div style={{ flex: 1 }} />
        <Space>
          <Button icon={<IconClose />} onClick={onClose}>取消</Button>
          <Button type="primary" icon={<IconSave />} onClick={handleSave}>保存工作流</Button>
        </Space>
      </div>

      {/* ── Main area ── */}
      <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        {/* Left palette */}
        <div style={{
          width: 164, flexShrink: 0, borderRight: '1px solid #E5E6EB',
          padding: '12px 10px', display: 'flex', flexDirection: 'column', gap: 8, overflowY: 'auto',
        }}>
          <div style={{ fontSize: 11, color: '#86909C', fontWeight: 600, letterSpacing: '0.05em', marginBottom: 2 }}>节点类型</div>
          {(Object.entries(NODE_META) as [WFNodeType, typeof NODE_META[WFNodeType]][]).map(([type, meta]) => (
            <button key={type} onClick={() => addNode(type)} style={{
              display: 'flex', alignItems: 'center', gap: 8, padding: '8px 10px',
              borderRadius: 8, border: `1.5px solid ${meta.color}22`,
              background: meta.bg, cursor: 'pointer', textAlign: 'left', width: '100%',
              transition: 'box-shadow .15s, transform .1s',
            }}
              onMouseEnter={e => { (e.currentTarget as HTMLElement).style.boxShadow = `0 2px 8px ${meta.color}33`; (e.currentTarget as HTMLElement).style.transform = 'scale(1.02)'; }}
              onMouseLeave={e => { (e.currentTarget as HTMLElement).style.boxShadow = 'none'; (e.currentTarget as HTMLElement).style.transform = 'scale(1)'; }}
            >
              <span style={{ fontSize: 18 }}>{meta.icon}</span>
              <span style={{ fontSize: 12, fontWeight: 600, color: meta.color }}>{meta.label}</span>
            </button>
          ))}
          <div style={{ marginTop: 'auto', padding: '8px 4px', fontSize: 11, color: '#86909C', lineHeight: 1.6 }}>
            <div>• 点击节点类型添加到画布</div>
            <div>• 拖动节点手柄连接</div>
            <div>• 点击节点配置参数</div>
            <div>• Delete 键删除选中节点</div>
          </div>
        </div>

        {/* Canvas */}
        <div ref={canvasRef} style={{ flex: 1, position: 'relative' }}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            nodeTypes={NODE_TYPES}
            onNodeClick={(_, node) => setSelectedNodeId(node.id)}
            onPaneClick={() => setSelectedNodeId(null)}
            fitView
            fitViewOptions={{ padding: 0.3 }}
            deleteKeyCode={null}
            defaultEdgeOptions={{ animated: true, style: { stroke: '#165DFF', strokeWidth: 2 } }}
          >
            <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#E5E6EB" />
            <Controls />
            {nodes.length === 0 && (
              <div style={{
                position: 'absolute', inset: 0, display: 'flex', alignItems: 'center',
                justifyContent: 'center', pointerEvents: 'none', zIndex: 5,
              }}>
                <div style={{ textAlign: 'center', color: '#C9CDD4' }}>
                  <IconDragDot style={{ fontSize: 40, marginBottom: 8 }} />
                  <div style={{ fontSize: 15 }}>从左侧点击节点类型开始构建工作流</div>
                </div>
              </div>
            )}
          </ReactFlow>
        </div>

        {/* Right params panel */}
        {selectedNode && (
          <div style={{
            width: 240, flexShrink: 0, borderLeft: '1px solid #E5E6EB',
            overflowY: 'auto', background: '#FAFAFA',
          }}>
            <ParamsPanel
              node={selectedNode}
              onChange={updateNodeParams}
              onDelete={deleteNode}
            />
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Public API ───────────────────────────────────────────────────────────────
interface WorkflowEditorProps {
  visible: boolean;
  initialGraphJSON?: string;
  initialName?: string;
  onSave: (name: string, graphJSON: string) => void;
  onClose: () => void;
}

export default function WorkflowEditor({ visible, initialGraphJSON, initialName = '', onSave, onClose }: WorkflowEditorProps) {
  const [name, setName] = useState(initialName);

  useEffect(() => { setName(initialName); }, [initialName]);

  if (!visible) return null;

  const { nodes: initNodes, edges: initEdges } = graphJSONToFlow(initialGraphJSON ?? '[]');

  const handleSave = (graphJSON: string) => {
    if (!name.trim()) { Message.warning('请先填写工作流名称'); return; }
    onSave(name.trim(), graphJSON);
  };

  return (
    <div style={{
      position: 'fixed', inset: 0, zIndex: 1050,
      display: 'flex', flexDirection: 'column',
    }}>
      <ReactFlowProvider>
        <EditorInner
          initialNodes={initNodes}
          initialEdges={initEdges}
          onSave={handleSave}
          onClose={onClose}
          workflowName={name}
          onNameChange={setName}
        />
      </ReactFlowProvider>
    </div>
  );
}
