import { useState, useCallback, useEffect, useRef } from 'react';
import {
  ReactFlow, ReactFlowProvider, addEdge, useNodesState, useEdgesState,
  Controls, Background, BackgroundVariant, Handle, Position, MarkerType,
  type Node, type Edge, type Connection, type NodeProps, type EdgeProps,
  useReactFlow, BaseEdge, EdgeLabelRenderer, getBezierPath,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Button, Space, Input, InputNumber, Select, Typography, Tooltip, Message } from '@arco-design/web-react';
import { IconSave, IconClose, IconDelete, IconPlus } from '@arco-design/web-react/icon';

// ─── Backend DAG types ─────────────────────────────────────────────────────────
type WFNodeType = 'start' | 'end' | 'acme_issue' | 'deploy_cert' | 'dns_validate' | 'notify' | 'condition' | 'wait';

interface WFNodeDef {
  id: string;
  type: WFNodeType;
  params?: Record<string, unknown>;
  depends_on?: string[];
  edge_labels?: Record<string, string>; // source_node_id → edge label
}

// ─── Custom SVG icons ──────────────────────────────────────────────────────────
const SvgCert = ({ c }: { c: string }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={c} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <rect x="2" y="6" width="12" height="9" rx="1.5" />
    <path d="M5 6V5a3 3 0 0 1 6 0v1" />
    <circle cx="8" cy="10.5" r="1.5" fill={c} stroke="none" />
    <path d="M8 12v1.5" />
  </svg>
);
const SvgDeploy = ({ c }: { c: string }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={c} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M8 2v9M5 5l3-3 3 3" />
    <path d="M2 12h4a1 1 0 0 1 1 1v1h2v-1a1 1 0 0 1 1-1h4" />
  </svg>
);
const SvgGlobe = ({ c }: { c: string }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={c} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="8" cy="8" r="6" />
    <path d="M8 2c-1.5 1.5-1.5 10.5 0 12M8 2c1.5 1.5 1.5 10.5 0 12" />
    <path d="M2 8h12M3 5h10M3 11h10" />
  </svg>
);
const SvgBell = ({ c }: { c: string }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={c} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M4.5 7a3.5 3.5 0 0 1 7 0v3l1.5 1.5H3L4.5 10V7z" />
    <path d="M6.5 13a1.5 1.5 0 0 0 3 0" />
  </svg>
);
const SvgDiamond = ({ c }: { c: string }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={c} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M8 2 L14 8 L8 14 L2 8 Z" />
    <path d="M5 8h6M8 5v6" />
  </svg>
);
const SvgClock = ({ c }: { c: string }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={c} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="8" cy="8" r="6" />
    <path d="M8 5v3.5l2 1.5" />
  </svg>
);
const SvgStart = ({ c }: { c: string }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill={c} stroke="none">
    <path d="M6 4l6 4-6 4z" />
  </svg>
);
const SvgEnd = ({ c }: { c: string }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={c} strokeWidth="1.5">
    <circle cx="8" cy="8" r="4" fill={c} />
    <circle cx="8" cy="8" r="6" />
  </svg>
);

// ─── Node meta ─────────────────────────────────────────────────────────────────
type IconFC = React.FC<{ c: string }>;
type FieldType = { key: string; label: string; type: 'text' | 'number' | 'select' | 'textarea'; options?: string[]; placeholder?: string };

const NODE_META: Record<WFNodeType, {
  label: string; color: string; bg: string; border: string; Icon: IconFC;
  shape: 'rect' | 'diamond' | 'circle';
  palette: boolean; // show in palette
  fields: FieldType[];
}> = {
  start: {
    label: '开始', color: '#00B42A', bg: '#E8FFEE', border: '#8FEEA9',
    Icon: SvgStart, shape: 'circle', palette: true, fields: [],
  },
  end: {
    label: '结束', color: '#86909C', bg: '#F2F3F5', border: '#C9CDD4',
    Icon: SvgEnd, shape: 'circle', palette: true, fields: [],
  },
  condition: {
    label: '条件判断', color: '#D4A400', bg: '#FFFAE8', border: '#FFE58F',
    Icon: SvgDiamond, shape: 'diamond', palette: true,
    fields: [
      { key: 'label', label: '节点名称', type: 'text', placeholder: '如：检查证书是否过期' },
      { key: 'expression', label: '条件表达式', type: 'text', placeholder: 'input.status == "success"' },
    ],
  },
  acme_issue: {
    label: '申请证书', color: '#165DFF', bg: '#EAF0FF', border: '#AAC4FF',
    Icon: SvgCert, shape: 'rect', palette: true,
    fields: [
      { key: 'domain_name', label: '域名', type: 'text', placeholder: 'example.com' },
      { key: 'acme_provider', label: 'ACME 提供商', type: 'select', options: ['letsencrypt', 'zerossl'] },
      { key: 'challenge', label: '验证方式', type: 'select', options: ['dns-01'] },
      { key: 'dns_account_id', label: 'DNS 账户 ID', type: 'number' },
      { key: 'account_email', label: '联系邮箱', type: 'text', placeholder: 'admin@example.com' },
      { key: 'staging', label: '测试模式', type: 'select', options: ['false', 'true'] },
    ],
  },
  deploy_cert: {
    label: '部署证书', color: '#7816FF', bg: '#F4E8FF', border: '#D3AEFF',
    Icon: SvgDeploy, shape: 'rect', palette: true,
    fields: [
      { key: 'cert_source_node', label: '来源节点 ID（留空取上游）', type: 'text', placeholder: '' },
      { key: 'target_node_id', label: '目标 CDN 节点 ID', type: 'number' },
    ],
  },
  dns_validate: {
    label: '检查 DNS', color: '#FF7D00', bg: '#FFF3E8', border: '#FFCB8F',
    Icon: SvgGlobe, shape: 'rect', palette: true,
    fields: [
      { key: 'domain_name', label: '域名', type: 'text', placeholder: 'example.com' },
      { key: 'record_type', label: '记录类型', type: 'select', options: ['A', 'AAAA', 'CNAME', 'TXT', 'MX'] },
      { key: 'expected_value', label: '期望值', type: 'text', placeholder: 'IP 或目标值' },
      { key: 'timeout_seconds', label: '超时（秒）', type: 'number' },
    ],
  },
  notify: {
    label: '发送通知', color: '#00B42A', bg: '#E8FFEE', border: '#8FEEA9',
    Icon: SvgBell, shape: 'rect', palette: true,
    fields: [
      { key: 'channel_id', label: '通知渠道 ID', type: 'number' },
      { key: 'message', label: '消息内容', type: 'textarea', placeholder: '工作流已完成' },
    ],
  },
  wait: {
    label: '等待', color: '#577590', bg: '#EEF3F7', border: '#B0C4D8',
    Icon: SvgClock, shape: 'rect', palette: true,
    fields: [
      { key: 'seconds', label: '等待时长（秒）', type: 'number' },
    ],
  },
};

// ─── Custom edge with labeled midpoint ────────────────────────────────────────
function LabeledEdge({
  id, sourceX, sourceY, targetX, targetY,
  sourcePosition, targetPosition,
  data, selected,
}: EdgeProps) {
  const [edgePath, labelX, labelY] = getBezierPath({ sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition });
  const label = (data as Record<string, unknown>)?.label as string | undefined;
  return (
    <>
      <BaseEdge id={id} path={edgePath} style={{ stroke: selected ? '#165DFF' : '#9CA3AF', strokeWidth: selected ? 2 : 1.5 }} markerEnd={`url(#arrow-${selected ? 'blue' : 'gray'})`} />
      {label && (
        <EdgeLabelRenderer>
          <div
            style={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              pointerEvents: 'all',
              background: '#fff',
              border: '1px solid #E5E6EB',
              borderRadius: 4,
              padding: '2px 7px',
              fontSize: 11,
              color: '#4E5969',
              fontWeight: 500,
              boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
              whiteSpace: 'nowrap',
            }}
            className="nodrag nopan"
          >
            {label}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
}

// ─── Rect node ────────────────────────────────────────────────────────────────
function RectNode({ id, data, selected }: NodeProps) {
  const meta = NODE_META[data.nodeType as WFNodeType];
  if (!meta) return null;
  const { Icon } = meta;
  const hint = (data.params as Record<string, unknown>)?.domain_name as string
    || (data.params as Record<string, unknown>)?.label as string
    || (data.params as Record<string, unknown>)?.message as string
    || null;
  return (
    <div style={{
      minWidth: 148,
      background: selected ? meta.bg : '#fff',
      border: `1.5px solid ${selected ? meta.color : meta.border}`,
      borderRadius: 8,
      boxShadow: selected ? `0 0 0 3px ${meta.color}28, 0 4px 12px ${meta.color}18` : '0 2px 8px rgba(0,0,0,0.07)',
      transition: 'all .15s',
    }}>
      <Handle type="target" position={Position.Left}
        style={{ background: '#fff', border: `2px solid ${meta.color}`, width: 10, height: 10 }} />
      <div style={{
        padding: '5px 11px',
        borderBottom: `1px solid ${selected ? meta.border : '#F0F1F5'}`,
        display: 'flex', alignItems: 'center', gap: 6,
        background: selected ? `${meta.color}18` : `${meta.color}0C`,
        borderRadius: '7px 7px 0 0',
      }}>
        <Icon c={meta.color} />
        <span style={{ fontSize: 12, fontWeight: 700, color: meta.color }}>{meta.label}</span>
      </div>
      <div style={{ padding: '5px 11px 7px' }}>
        <div style={{ fontSize: 11, color: hint ? '#4E5969' : '#C9CDD4', maxWidth: 128, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {hint ?? '点击配置参数'}
        </div>
        <div style={{ fontSize: 10, color: '#D4D8E0', fontFamily: 'monospace', marginTop: 2 }}>{id}</div>
      </div>
      <Handle type="source" position={Position.Right}
        style={{ background: meta.color, border: '2px solid #fff', width: 10, height: 10 }} />
    </div>
  );
}

// ─── Diamond node ─────────────────────────────────────────────────────────────
const DW = 170, DH = 90;
function DiamondNode({ id, data, selected }: NodeProps) {
  const meta = NODE_META[data.nodeType as WFNodeType];
  if (!meta) return null;
  const pts = `${DW / 2},4 ${DW - 4},${DH / 2} ${DW / 2},${DH - 4} 4,${DH / 2}`;
  const lbl = (data.params as Record<string, unknown>)?.label as string || meta.label;
  return (
    <div style={{ width: DW, height: DH, position: 'relative' }}>
      <Handle type="target" position={Position.Left}
        style={{ background: '#fff', border: `2px solid ${meta.color}`, width: 10, height: 10, left: -1, top: DH / 2 }} />
      <svg width={DW} height={DH} style={{ position: 'absolute', top: 0, left: 0, overflow: 'visible' }}>
        <defs>
          <filter id={`shadow-d-${id}`} x="-20%" y="-20%" width="140%" height="140%">
            <feDropShadow dx="0" dy="2" stdDeviation="3" floodColor={meta.color} floodOpacity={selected ? '0.18' : '0.06'} />
          </filter>
        </defs>
        <polygon
          points={pts}
          fill={selected ? meta.bg : '#fff'}
          stroke={selected ? meta.color : meta.border}
          strokeWidth={selected ? '2' : '1.5'}
          filter={`url(#shadow-d-${id})`}
        />
      </svg>
      <div style={{
        position: 'absolute', inset: 0,
        display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 4,
        pointerEvents: 'none',
      }}>
        <SvgDiamond c={meta.color} />
        <span style={{
          fontSize: 11, fontWeight: 600, color: meta.color,
          maxWidth: DW * 0.6, textAlign: 'center', lineHeight: 1.3,
          overflow: 'hidden', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical',
        }}>{lbl}</span>
      </div>
      {/* Right (true / default output) */}
      <Handle type="source" position={Position.Right}
        style={{ background: meta.color, border: '2px solid #fff', width: 10, height: 10, right: -1, top: DH / 2 }} />
      {/* Bottom (false / alternate output) */}
      <Handle type="source" id="alt" position={Position.Bottom}
        style={{ background: meta.color, border: '2px solid #fff', width: 10, height: 10, bottom: -1, left: DW / 2 }} />
    </div>
  );
}

// ─── Circle node (start / end) ────────────────────────────────────────────────
const CR = 52;
function CircleNode({ id: _id, data, selected }: NodeProps) {
  const meta = NODE_META[data.nodeType as WFNodeType];
  if (!meta) return null;
  const isEnd = (data.nodeType as WFNodeType) === 'end';
  return (
    <div style={{
      width: CR, height: CR, borderRadius: '50%',
      background: selected ? meta.bg : '#fff',
      border: `2px solid ${selected ? meta.color : meta.border}`,
      display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column',
      boxShadow: selected ? `0 0 0 3px ${meta.color}30` : '0 2px 8px rgba(0,0,0,0.08)',
      transition: 'all .15s',
    }}>
      {!isEnd && (
        <Handle type="source" position={Position.Right}
          style={{ background: meta.color, border: '2px solid #fff', width: 10, height: 10 }} />
      )}
      {isEnd && (
        <Handle type="target" position={Position.Left}
          style={{ background: '#fff', border: `2px solid ${meta.color}`, width: 10, height: 10 }} />
      )}
      {/* Also allow target on start (for loops), source on end */}
      {!isEnd && (
        <Handle type="target" position={Position.Left}
          style={{ background: '#fff', border: `2px solid ${meta.color}`, width: 8, height: 8, opacity: 0.4 }} />
      )}
      <meta.Icon c={meta.color} />
      <span style={{ fontSize: 11, fontWeight: 700, color: meta.color, marginTop: 2, lineHeight: 1 }}>{meta.label}</span>
    </div>
  );
}

const RF_NODE_TYPES = { rectNode: RectNode, diamondNode: DiamondNode, circleNode: CircleNode };
const EDGE_TYPES = { labeled: LabeledEdge };

function getRFType(t: WFNodeType) {
  if (t === 'start' || t === 'end') return 'circleNode';
  if (t === 'condition') return 'diamondNode';
  return 'rectNode';
}

// ─── Layout ───────────────────────────────────────────────────────────────────
function autoLayout(defs: WFNodeDef[]): Node[] {
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
  const cols: Record<number, string[]> = {};
  Object.entries(depthMap).forEach(([id, d]) => { cols[d] = cols[d] ?? []; cols[d].push(id); });
  return defs.map(def => {
    const depth = depthMap[def.id] ?? 0;
    const col = cols[depth] ?? [];
    const row = col.indexOf(def.id);
    return {
      id: def.id,
      type: getRFType(def.type),
      position: { x: depth * 240 + 40, y: row * 160 + 40 },
      data: { nodeType: def.type, params: def.params ?? {} },
    };
  });
}

function graphJSONToFlow(json: string): { nodes: Node[]; edges: Edge[] } {
  try {
    const defs: WFNodeDef[] = JSON.parse(json);
    if (!Array.isArray(defs)) throw new Error();
    const nodes = autoLayout(defs);
    const edges: Edge[] = [];
    defs.forEach(def => {
      (def.depends_on ?? []).forEach(srcId => {
        const label = def.edge_labels?.[srcId];
        edges.push({
          id: `e-${srcId}-${def.id}`,
          source: srcId, target: def.id,
          type: 'labeled',
          data: label ? { label } : {},
          markerEnd: { type: MarkerType.ArrowClosed, color: '#9CA3AF', width: 14, height: 14 },
        });
      });
    });
    return { nodes, edges };
  } catch {
    return { nodes: [], edges: [] };
  }
}

function flowToGraphJSON(nodes: Node[], edges: Edge[]): string {
  return JSON.stringify(nodes.map(n => {
    const incoming = edges.filter(e => e.target === n.id);
    const edge_labels: Record<string, string> = {};
    incoming.forEach(e => { if ((e.data as Record<string, unknown>)?.label) edge_labels[e.source] = String((e.data as Record<string, unknown>).label); });
    return {
      id: n.id,
      type: n.data.nodeType as WFNodeType,
      params: (n.data.params as Record<string, unknown>) ?? {},
      depends_on: incoming.map(e => e.source),
      ...(Object.keys(edge_labels).length > 0 ? { edge_labels } : {}),
    } satisfies WFNodeDef;
  }));
}

// ─── Params panel ──────────────────────────────────────────────────────────────
function NodePanel({ node, onChange, onDelete }: {
  node: Node; onChange: (id: string, params: Record<string, unknown>) => void; onDelete: (id: string) => void;
}) {
  const meta = NODE_META[node.data.nodeType as WFNodeType];
  const params = (node.data.params ?? {}) as Record<string, unknown>;
  const { Icon } = meta;
  const set = (k: string, v: unknown) => onChange(node.id, { ...params, [k]: v });
  if (!meta) return null;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ padding: '12px 14px', borderBottom: '1px solid #F0F1F5', background: `${meta.color}09`, flexShrink: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{ width: 30, height: 30, borderRadius: 8, background: meta.bg, border: `1.5px solid ${meta.border}`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
            <Icon c={meta.color} />
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontWeight: 700, fontSize: 13, color: meta.color }}>{meta.label}</div>
            <div style={{ fontSize: 10, color: '#86909C', fontFamily: 'monospace', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{node.id}</div>
          </div>
          <Tooltip content="删除节点">
            <Button size="mini" type="text" status="danger" icon={<IconDelete />} onClick={() => onDelete(node.id)} />
          </Tooltip>
        </div>
      </div>
      <div style={{ flex: 1, overflow: 'auto', padding: '12px 14px', display: 'flex', flexDirection: 'column', gap: 10 }}>
        {meta.fields.map(f => (
          <div key={f.key}>
            <div style={{ fontSize: 12, color: '#4E5969', marginBottom: 4, fontWeight: 600 }}>{f.label}</div>
            {f.type === 'select' ? (
              <Select size="small" value={params[f.key] as string ?? ''} onChange={v => set(f.key, v)} style={{ width: '100%' }}>
                {(f.options ?? []).map(o => <Select.Option key={o} value={o}>{o}</Select.Option>)}
              </Select>
            ) : f.type === 'number' ? (
              <InputNumber size="small" value={params[f.key] as number ?? undefined} onChange={v => set(f.key, v)} style={{ width: '100%' }} placeholder="0" />
            ) : f.type === 'textarea' ? (
              <Input.TextArea rows={3} value={params[f.key] as string ?? ''} onChange={v => set(f.key, v)} placeholder={f.placeholder} style={{ fontSize: 12 }} />
            ) : (
              <Input size="small" value={params[f.key] as string ?? ''} onChange={v => set(f.key, v)} placeholder={f.placeholder} />
            )}
          </div>
        ))}
        {meta.fields.length === 0 && <div style={{ color: '#C9CDD4', fontSize: 12, textAlign: 'center', paddingTop: 16 }}>此节点无可配置参数</div>}
      </div>
    </div>
  );
}

function EdgePanel({ edge, onUpdate, onDelete }: {
  edge: Edge; onUpdate: (id: string, label: string) => void; onDelete: (id: string) => void;
}) {
  const [val, setVal] = useState<string>((edge.data as Record<string, unknown>)?.label as string ?? '');
  return (
    <div style={{ padding: '14px', display: 'flex', flexDirection: 'column', gap: 10 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 2 }}>
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M2 8h12M10 5l4 3-4 3" stroke="#86909C" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" /></svg>
        <Typography.Text bold style={{ fontSize: 13, color: '#1D2129' }}>连接属性</Typography.Text>
        <div style={{ flex: 1 }} />
        <Tooltip content="删除连接">
          <Button size="mini" type="text" status="danger" icon={<IconDelete />} onClick={() => onDelete(edge.id)} />
        </Tooltip>
      </div>
      <div style={{ fontSize: 11, color: '#86909C', background: '#F7F8FA', borderRadius: 6, padding: '4px 8px', fontFamily: 'monospace' }}>
        {edge.source} → {edge.target}
      </div>
      <div>
        <div style={{ fontSize: 12, color: '#4E5969', marginBottom: 4, fontWeight: 600 }}>分支标签（可选）</div>
        <Input
          size="small"
          value={val}
          onChange={setVal}
          placeholder="如：完整 / 未申请 / 已过期"
          onBlur={() => onUpdate(edge.id, val)}
          onPressEnter={() => onUpdate(edge.id, val)}
        />
        <div style={{ fontSize: 11, color: '#C9CDD4', marginTop: 4 }}>条件节点的输出分支建议填写标签</div>
      </div>
    </div>
  );
}

// ─── Palette button ───────────────────────────────────────────────────────────
function PaletteBtn({ type, onClick }: { type: WFNodeType; onClick: () => void }) {
  const meta = NODE_META[type];
  const [hov, setHov] = useState(false);
  return (
    <button
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        display: 'flex', alignItems: 'center', gap: 9, padding: '8px 10px',
        borderRadius: 8, border: `1.5px solid ${hov ? meta.color : meta.border}`,
        background: hov ? meta.bg : '#fff', cursor: 'pointer', textAlign: 'left', width: '100%',
        transition: 'all .15s', boxShadow: hov ? `0 2px 8px ${meta.color}22` : 'none',
      }}
    >
      <div style={{
        width: 26, height: 26, borderRadius: 6, flexShrink: 0,
        background: hov ? meta.color : meta.bg,
        border: `1.5px solid ${hov ? meta.color : meta.border}`,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        transition: 'all .15s',
      }}>
        <meta.Icon c={hov ? '#fff' : meta.color} />
      </div>
      <span style={{ fontSize: 12, fontWeight: 600, color: hov ? meta.color : '#1D2129', flex: 1 }}>{meta.label}</span>
      <IconPlus style={{ fontSize: 11, color: hov ? meta.color : '#C9CDD4' }} />
    </button>
  );
}

// ─── Inner editor ─────────────────────────────────────────────────────────────
let nodeCounter = 1;
const newId = () => `n${Date.now()}_${nodeCounter++}`;

type SelType = { kind: 'node'; id: string } | { kind: 'edge'; id: string } | null;

function EditorInner({ initialNodes, initialEdges, onSave, onClose, workflowName, onNameChange }: {
  initialNodes: Node[]; initialEdges: Edge[];
  onSave: (json: string) => void; onClose: () => void;
  workflowName: string; onNameChange: (v: string) => void;
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [sel, setSel] = useState<SelType>(null);
  const { screenToFlowPosition } = useReactFlow();
  const canvasRef = useRef<HTMLDivElement>(null);

  const selNode = sel?.kind === 'node' ? nodes.find(n => n.id === sel.id) ?? null : null;
  const selEdge = sel?.kind === 'edge' ? edges.find(e => e.id === sel.id) ?? null : null;

  const onConnect = useCallback((conn: Connection) =>
    setEdges(eds => addEdge({
      ...conn, type: 'labeled', data: {},
      markerEnd: { type: MarkerType.ArrowClosed, color: '#9CA3AF', width: 14, height: 14 },
    }, eds)), [setEdges]);

  const addNode = (type: WFNodeType) => {
    const rect = canvasRef.current?.getBoundingClientRect();
    const pos = screenToFlowPosition({ x: (rect?.left ?? 0) + (rect?.width ?? 600) / 2, y: (rect?.top ?? 0) + (rect?.height ?? 400) / 2 });
    const id = newId();
    setNodes(nds => [...nds, { id, type: getRFType(type), position: { x: pos.x + (Math.random() - 0.5) * 100, y: pos.y + (Math.random() - 0.5) * 60 }, data: { nodeType: type, params: {} } }]);
    setSel({ kind: 'node', id });
  };

  const updateNodeParams = (id: string, params: Record<string, unknown>) =>
    setNodes(nds => nds.map(n => n.id === id ? { ...n, data: { ...n.data, params } } : n));

  const updateEdgeLabel = (id: string, label: string) =>
    setEdges(eds => eds.map(e => e.id === id ? { ...e, data: { label: label || undefined } } : e));

  const deleteNode = (id: string) => {
    setNodes(nds => nds.filter(n => n.id !== id));
    setEdges(eds => eds.filter(e => e.source !== id && e.target !== id));
    setSel(null);
  };

  const deleteEdge = (id: string) => {
    setEdges(eds => eds.filter(e => e.id !== id));
    setSel(null);
  };

  useEffect(() => {
    const h = (e: KeyboardEvent) => {
      if ((e.key === 'Delete' || e.key === 'Backspace') && !(e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) {
        if (sel?.kind === 'node') deleteNode(sel.id);
        if (sel?.kind === 'edge') deleteEdge(sel.id);
      }
    };
    window.addEventListener('keydown', h);
    return () => window.removeEventListener('keydown', h);
  }, [sel]); // eslint-disable-line

  const handleSave = () => {
    if (nodes.length === 0) { Message.warning('至少需要一个节点'); return; }
    onSave(flowToGraphJSON(nodes, edges));
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', background: '#fff' }}>
      {/* Top bar */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', height: 54, flexShrink: 0, background: '#fff', borderBottom: '1px solid #E5E6EB', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
          <rect x="1" y="4" width="18" height="12" rx="3" fill="#165DFF" opacity=".12" stroke="#165DFF" strokeWidth="1.5" />
          <circle cx="6" cy="10" r="2.5" fill="#165DFF" />
          <path d="M11 10h5M11 7.5h3M11 12.5h3" stroke="#165DFF" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
        <Typography.Text bold style={{ fontSize: 15, color: '#1D2129', whiteSpace: 'nowrap' }}>工作流编辑器</Typography.Text>
        <div style={{ width: 1, height: 18, background: '#E5E6EB' }} />
        <Input value={workflowName} onChange={onNameChange} placeholder="输入工作流名称（必填）"
          prefix={<span style={{ color: '#86909C', fontSize: 12 }}>名称</span>} style={{ width: 240 }} />
        <div style={{ flex: 1 }} />
        <span style={{ fontSize: 11, color: '#86909C' }}>{nodes.length} 节点 · {edges.length} 连接</span>
        <Space>
          <Button icon={<IconClose />} onClick={onClose}>取消</Button>
          <Button type="primary" icon={<IconSave />} onClick={handleSave}>保存工作流</Button>
        </Space>
      </div>

      {/* Main */}
      <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        {/* Left palette */}
        <div style={{ width: 182, flexShrink: 0, background: '#fff', borderRight: '1px solid #E5E6EB', display: 'flex', flexDirection: 'column' }}>
          <div style={{ padding: '10px 12px 4px' }}>
            <span style={{ fontSize: 11, fontWeight: 700, color: '#86909C', letterSpacing: '0.08em', textTransform: 'uppercase' }}>节点类型</span>
          </div>
          <div style={{ flex: 1, overflowY: 'auto', padding: '4px 10px 10px', display: 'flex', flexDirection: 'column', gap: 4 }}>
            {/* Flow control */}
            <div style={{ fontSize: 10, color: '#C9CDD4', fontWeight: 600, padding: '4px 2px 2px', textTransform: 'uppercase', letterSpacing: '0.06em' }}>流程控制</div>
            {(['start', 'end', 'condition', 'wait'] as WFNodeType[]).map(t => <PaletteBtn key={t} type={t} onClick={() => addNode(t)} />)}
            <div style={{ fontSize: 10, color: '#C9CDD4', fontWeight: 600, padding: '8px 2px 2px', textTransform: 'uppercase', letterSpacing: '0.06em' }}>操作节点</div>
            {(['acme_issue', 'deploy_cert', 'dns_validate', 'notify'] as WFNodeType[]).map(t => <PaletteBtn key={t} type={t} onClick={() => addNode(t)} />)}
          </div>
          <div style={{ padding: '10px 12px', borderTop: '1px solid #F0F1F5', background: '#FAFBFC' }}>
            <div style={{ fontSize: 10, color: '#86909C', lineHeight: 1.8 }}>
              <div>• 点击添加节点到画布</div>
              <div>• 拖动连接点连线</div>
              <div>• 点击连线添加分支标签</div>
              <div>• <kbd style={{ background: '#E5E6EB', padding: '0 3px', borderRadius: 3, fontSize: 9 }}>Del</kbd> 删除选中元素</div>
            </div>
          </div>
        </div>

        {/* Canvas */}
        <div ref={canvasRef} style={{ flex: 1, position: 'relative' }}>
          {/* SVG arrow markers */}
          <svg style={{ position: 'absolute', width: 0, height: 0 }}>
            <defs>
              <marker id="arrow-gray" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                <path d="M0,0 L0,6 L8,3 z" fill="#9CA3AF" />
              </marker>
              <marker id="arrow-blue" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                <path d="M0,0 L0,6 L8,3 z" fill="#165DFF" />
              </marker>
            </defs>
          </svg>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            nodeTypes={RF_NODE_TYPES}
            edgeTypes={EDGE_TYPES}
            onNodeClick={(_, node) => setSel({ kind: 'node', id: node.id })}
            onEdgeClick={(_, edge) => setSel({ kind: 'edge', id: edge.id })}
            onPaneClick={() => setSel(null)}
            fitView fitViewOptions={{ padding: 0.35 }}
            deleteKeyCode={null}
            defaultEdgeOptions={{ type: 'labeled', data: {}, markerEnd: { type: MarkerType.ArrowClosed, color: '#9CA3AF', width: 14, height: 14 } }}
            style={{ background: '#fff' }}
          >
            <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#E5E7EB" />
            <Controls style={{ bottom: 20, left: 12 }} />
            {nodes.length === 0 && (
              <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', pointerEvents: 'none', zIndex: 5 }}>
                <div style={{ textAlign: 'center', color: '#C9CDD4' }}>
                  <svg width="64" height="52" viewBox="0 0 64 52" fill="none" style={{ marginBottom: 14 }}>
                    <rect x="4" y="4" width="56" height="44" rx="8" fill="#F9FAFB" stroke="#E5E6EB" strokeWidth="1.5" />
                    {/* Start circle */}
                    <circle cx="16" cy="26" r="7" fill="#E5E6EB" />
                    {/* Arrow */}
                    <path d="M23 26h8" stroke="#E5E6EB" strokeWidth="1.5" strokeDasharray="2 2" />
                    {/* Diamond */}
                    <polygon points="37,20 44,26 37,32 30,26" fill="#F3F4F6" stroke="#E5E6EB" strokeWidth="1.5" />
                    <path d="M44 26h6" stroke="#E5E6EB" strokeWidth="1.5" strokeDasharray="2 2" />
                    {/* End circle */}
                    <circle cx="56" cy="26" r="4" fill="#E5E6EB" />
                  </svg>
                  <div style={{ fontSize: 15, fontWeight: 600, color: '#86909C', marginBottom: 4 }}>从左侧添加节点开始构建</div>
                  <div style={{ fontSize: 12, color: '#C9CDD4' }}>推荐结构：开始 → 条件判断 → 操作节点 → 结束</div>
                </div>
              </div>
            )}
          </ReactFlow>
        </div>

        {/* Right panel */}
        <div style={{ width: 252, flexShrink: 0, borderLeft: '1px solid #E5E6EB', background: selNode || selEdge ? '#fff' : '#FAFBFC', overflowY: 'auto' }}>
          {selNode ? (
            <NodePanel node={selNode} onChange={updateNodeParams} onDelete={deleteNode} />
          ) : selEdge ? (
            <EdgePanel edge={selEdge} onUpdate={updateEdgeLabel} onDelete={deleteEdge} />
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', padding: 20 }}>
              <div style={{ textAlign: 'center', color: '#C9CDD4' }}>
                <svg width="36" height="36" viewBox="0 0 36 36" fill="none" style={{ marginBottom: 8 }}>
                  <circle cx="18" cy="18" r="13" fill="#F2F3F5" stroke="#E5E6EB" strokeWidth="1.5" />
                  <path d="M18 12v6l4 3" stroke="#C9CDD4" strokeWidth="1.8" strokeLinecap="round" />
                </svg>
                <div style={{ fontSize: 12, lineHeight: 1.6 }}>点击节点或连线<br />查看和配置属性</div>
              </div>
            </div>
          )}
        </div>
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

  return (
    <div style={{ position: 'fixed', inset: 0, zIndex: 1050, display: 'flex', flexDirection: 'column' }}>
      <ReactFlowProvider>
        <EditorInner
          initialNodes={initNodes} initialEdges={initEdges}
          onSave={json => { if (!name.trim()) { Message.warning('请先填写工作流名称'); return; } onSave(name.trim(), json); }}
          onClose={onClose} workflowName={name} onNameChange={setName}
        />
      </ReactFlowProvider>
    </div>
  );
}
