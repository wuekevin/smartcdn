import {
  Button, Card, Table, Tag, Typography, Space, Popconfirm, Message,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import client from '../../api/client';

interface Domain {
  id: number;
  name: string;
  status: string;
  icp_filed: boolean;
  dns_account_id: number;
  created_at: string;
}

export default function DomainsPage() {
  const qc = useQueryClient();
  const [afterID] = useState(0);

  const { data, isLoading } = useQuery({
    queryKey: ['domains', afterID],
    queryFn: () =>
      client.get('/domains', { params: { after_id: afterID || undefined, limit: 20 } }).then((r) => r.data),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/domains/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['domains'] }); Message.success('域名已删除'); },
    onError: () => Message.error('删除失败'),
  });

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
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: Domain) => (
        <Space>
          <Popconfirm title="确认删除此域名？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>域名列表</Typography.Title>}
      extra={
        <Space>
          <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['domains'] })} />
          <Button type="primary" icon={<IconPlus />}>添加域名</Button>
        </Space>
      }
    >
      <Table
        loading={isLoading}
        columns={columns}
        data={data?.items ?? []}
        rowKey="id"
        pagination={false}
      />
    </Card>
  );
}
