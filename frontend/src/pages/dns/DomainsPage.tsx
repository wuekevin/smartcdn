import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Space, Popconfirm, Message, Modal, Form, Input, Switch, Select,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface Domain {
  id: number; name: string; status: string; icp_filed: boolean; is_active: boolean;
  dns_account_id: number; max_publish_ips: number; created_at: string;
}
interface DNSAccount { id: number; name: string; }

export default function DomainsPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<Domain | null>(null);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['domains'],
    queryFn: () => client.get('/domains', { params: { limit: 20 } }).then(r => r.data),
  });

  const { data: accountsData } = useQuery({
    queryKey: ['dns-accounts'],
    queryFn: () => client.get('/dns-accounts').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/domains', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['domains'] });
      Message.success('域名已添加');
      setVisible(false);
      form.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '添加失败');
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, ...body }: { id: number } & Record<string, unknown>) =>
      client.put(`/domains/${id}`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['domains'] });
      Message.success('已更新');
      setEditVisible(false);
      editForm.resetFields();
      setEditingRow(null);
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '更新失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/domains/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['domains'] }); Message.success('域名已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const accounts: DNSAccount[] = accountsData?.items ?? [];

  const handleCreate = async () => {
    const vals = await form.validate();
    // Ensure account_id is a number regardless of Select's internal type
    createMut.mutate({ ...vals, account_id: Number(vals.account_id) });
  };

  const handleEdit = (row: Domain) => {
    setEditingRow(row);
    editForm.setFieldsValue({
      icp_filed: row.icp_filed,
      is_active: row.is_active,
      max_publish_ips: row.max_publish_ips ?? 1,
    });
    setEditVisible(true);
  };

  const handleUpdate = async () => {
    if (!editingRow) return;
    const vals = await editForm.validate();
    updateMut.mutate({ id: editingRow.id, ...vals });
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '域名', dataIndex: 'name' },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={v === 'active' ? 'green' : 'orange'}>{v}</Tag>,
    },
    {
      title: 'ICP 备案', dataIndex: 'icp_filed',
      render: (v: boolean) => <Tag color={v ? 'arcoblue' : 'gray'}>{v ? '已备案' : '未备案'}</Tag>,
    },
    {
      title: '启用', dataIndex: 'is_active',
      render: (v: boolean) => <Tag color={v ? 'green' : 'gray'}>{v ? '是' : '否'}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: Domain) => (
        <Space>
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => handleEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除此域名？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>域名列表</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['domains'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>添加域名</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      {/* 新建弹窗 */}
      <Modal
        title="添加域名"
        visible={visible}
        onOk={handleCreate}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <Form.Item label="域名" field="name" rules={[{ required: true, message: '请输入域名' }]}>
            <Input placeholder="example.com" />
          </Form.Item>
          <Form.Item label="DNS 账户" field="account_id" rules={[{ required: true, message: '请选择 DNS 账户' }]}>
            <Select placeholder="选择 DNS 账户">
              {accounts.map(a => <Select.Option key={a.id} value={Number(a.id)}>{a.name}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item label="ICP 备案" field="icp_filed" triggerPropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑弹窗 */}
      <Modal
        title="编辑域名"
        visible={editVisible}
        onOk={handleUpdate}
        confirmLoading={updateMut.isPending}
        onCancel={() => { setEditVisible(false); editForm.resetFields(); setEditingRow(null); }}
        unmountOnExit
      >
        <Form form={editForm} layout="vertical">
          <Form.Item label="ICP 备案" field="icp_filed" triggerPropName="checked">
            <Switch checkedText="已备案" uncheckedText="未备案" />
          </Form.Item>
          <Form.Item label="启用状态" field="is_active" triggerPropName="checked">
            <Switch checkedText="启用" uncheckedText="禁用" />
          </Form.Item>
          <Form.Item label="最大发布 IP 数" field="max_publish_ips">
            <Input type="number" placeholder="1" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
