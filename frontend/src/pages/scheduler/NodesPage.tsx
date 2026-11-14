import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Space, Message, Modal, Form, Input, InputNumber,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface CDNNode {
  id: number;
  name: string;
  ip: string;
  region: string;
  isp: string;
  status: string;
  https_latency_ms: number;
  https_status_code: number;
  last_seen_at: string | null;
}

export default function NodesPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [form] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['cdn-nodes'],
    queryFn: () => client.get('/nodes').then((r) => r.data),
    refetchInterval: 30_000,
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/nodes', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['cdn-nodes'] });
      Message.success('节点已添加');
      setVisible(false);
      form.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '添加失败');
    },
  });

  const statusColor: Record<string, string> = {
    active: 'green', inactive: 'gray', maintenance: 'orange', offline: 'red',
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '节点名称', dataIndex: 'name' },
    { title: 'IP', dataIndex: 'ip', width: 150 },
    { title: '地区', dataIndex: 'region' },
    { title: 'ISP', dataIndex: 'isp' },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={statusColor[v] ?? 'gray'}>{v}</Tag>,
    },
    {
      title: 'HTTPS 延迟',
      render: (_: unknown, row: CDNNode) =>
        row.https_latency_ms > 0
          ? <span style={{ color: row.https_latency_ms > 300 ? 'orange' : 'green' }}>{row.https_latency_ms} ms</span>
          : <Tag color="gray">-</Tag>,
    },
    { title: '最后在线', dataIndex: 'last_seen_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: CDNNode) => (
        <Space>
          <Button size="mini" onClick={() =>
            client.post(`/nodes/${row.id}/https-check`).then(() => {
              qc.invalidateQueries({ queryKey: ['cdn-nodes'] });
              Message.success('HTTPS 检测已触发');
            })
          }>HTTPS 检测</Button>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>CDN 节点</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['cdn-nodes'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>添加节点</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="添加 CDN 节点"
        visible={visible}
        onOk={() => form.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <Form.Item label="节点名称" field="name" rules={[{ required: true, message: '请输入节点名称' }]}>
            <Input placeholder="如：北京-联通-01" />
          </Form.Item>
          <Form.Item label="IP 地址" field="ip" rules={[{ required: true, message: '请输入 IP 地址' }]}>
            <Input placeholder="1.2.3.4" />
          </Form.Item>
          <Form.Item label="地区" field="region">
            <Input placeholder="如：Beijing" />
          </Form.Item>
          <Form.Item label="ISP" field="isp">
            <Input placeholder="如：ChinaUnicom" />
          </Form.Item>
          <Form.Item label="源站地址" field="origin_addr">
            <Input placeholder="1.2.3.4 或域名" />
          </Form.Item>
          <Form.Item label="源站端口" field="origin_port" initialValue={80}>
            <InputNumber min={1} max={65535} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
