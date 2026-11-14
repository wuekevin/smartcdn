import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Space, Popconfirm, Message, Modal, Form, Input, Select, InputNumber,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit, IconSync } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface Domain { id: number; name: string; }
interface DNSRecord {
  id: number; host: string; type: string; value: string; ttl: number; line: string; is_external: boolean;
}

const RECORD_TYPES = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS', 'SRV', 'CAA'];

export default function DNSRecordsPage() {
  const qc = useQueryClient();
  const [domainId, setDomainId] = useState<number>(0);
  const [createVisible, setCreateVisible] = useState(false);
  const [editRecord, setEditRecord] = useState<DNSRecord | null>(null);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  const { data: domainsData } = useQuery({
    queryKey: ['domains-select'],
    queryFn: () => client.get('/domains', { params: { limit: 200 } }).then(r => r.data),
  });

  const { data, isLoading } = useQuery({
    queryKey: ['records', domainId],
    queryFn: () => client.get(`/domains/${domainId}/records`).then(r => r.data),
    enabled: domainId > 0,
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post(`/domains/${domainId}/records`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['records', domainId] });
      Message.success('记录已创建');
      setCreateVisible(false);
      form.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, body }: { id: number; body: unknown }) =>
      client.put(`/domains/${domainId}/records/${id}`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['records', domainId] });
      Message.success('记录已更新');
      setEditRecord(null);
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '更新失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/domains/${domainId}/records/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['records', domainId] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const syncMut = useMutation({
    mutationFn: () => client.post(`/domains/${domainId}/sync-from-provider`),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['records', domainId] });
      Message.success(`已从服务商同步 ${res.data?.count ?? 0} 条记录`);
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '同步失败');
    },
  });

  const domains: Domain[] = domainsData?.items ?? [];

  const recordForm = (f: ReturnType<typeof Form.useForm>[0]) => (
    <Form form={f} layout="vertical">
      <Form.Item label="主机记录" field="host" rules={[{ required: true }]}>
        <Input placeholder="@ 或 www 或 mail" />
      </Form.Item>
      <Form.Item label="记录类型" field="type" rules={[{ required: true }]}>
        <Select>
          {RECORD_TYPES.map(t => <Select.Option key={t} value={t}>{t}</Select.Option>)}
        </Select>
      </Form.Item>
      <Form.Item label="记录值" field="value" rules={[{ required: true }]}>
        <Input placeholder="IP 地址、域名或文本内容" />
      </Form.Item>
      <Form.Item label="TTL (秒)" field="ttl" initialValue={600}>
        <InputNumber min={1} max={86400} />
      </Form.Item>
      <Form.Item label="线路" field="line" initialValue="default">
        <Input placeholder="default" />
      </Form.Item>
    </Form>
  );

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '主机记录', dataIndex: 'host' },
    { title: '类型', dataIndex: 'type', width: 80 },
    { title: '值', dataIndex: 'value' },
    { title: 'TTL', dataIndex: 'ttl', width: 80 },
    { title: '线路', dataIndex: 'line', width: 100 },
    {
      title: '外部', dataIndex: 'is_external',
      render: (v: boolean) => <Tag color={v ? 'orange' : 'gray'}>{v ? '外部' : '托管'}</Tag>,
    },
    {
      title: '操作', render: (_: unknown, row: DNSRecord) => (
        <Space>
          <Button size="mini" icon={<IconEdit />} onClick={() => {
            setEditRecord(row);
            editForm.setFieldsValue({ host: row.host, type: row.type, value: row.value, ttl: row.ttl, line: row.line });
          }}>编辑</Button>
          <Popconfirm title="确认删除此记录？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>DNS 记录</Typography.Title>}
        extra={
          <Space>
            <Select
              placeholder="选择域名"
              style={{ width: 200 }}
              value={domainId || undefined}
              onChange={v => setDomainId(Number(v))}
            >
              {domains.map(d => <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>)}
            </Select>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['records', domainId] })} />
            <Popconfirm
              title="将从服务商拉取全部记录并覆盖本地，确认继续？"
              onOk={() => syncMut.mutate()}
              disabled={domainId === 0}
            >
              <Button icon={<IconSync />} disabled={domainId === 0} loading={syncMut.isPending}>从服务商同步</Button>
            </Popconfirm>
            <Button type="primary" icon={<IconPlus />} disabled={domainId === 0} onClick={() => setCreateVisible(true)}>
              添加记录
            </Button>
          </Space>
        }
      >
        <Table
          loading={isLoading}
          columns={columns}
          data={data?.items ?? []}
          rowKey="id"
          pagination={false}
          noDataElement={domainId === 0 ? <span>请先选择域名</span> : undefined}
        />
      </Card>

      <Modal
        title="添加 DNS 记录"
        visible={createVisible}
        onOk={() => form.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setCreateVisible(false); form.resetFields(); }}
        unmountOnExit
      >
        {recordForm(form)}
      </Modal>

      <Modal
        title="编辑 DNS 记录"
        visible={editRecord !== null}
        onOk={() => editForm.validate().then(vals => updateMut.mutate({ id: editRecord!.id, body: vals }))}
        confirmLoading={updateMut.isPending}
        onCancel={() => setEditRecord(null)}
        unmountOnExit
      >
        {recordForm(editForm)}
      </Modal>
    </>
  );
}
