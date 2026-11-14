import { Layout, Menu, Typography } from '@arco-design/web-react';
import {
  IconDashboard,
  IconCloud,
  IconApps,
  IconCompass,
  IconSafe,
  IconBulb,
  IconSettings,
} from '@arco-design/web-react/icon';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { useState } from 'react';

const { Sider, Header, Content } = Layout;
const { Item: MenuItem, SubMenu } = Menu;

const navItems = [
  { key: '/dashboard',                icon: <IconDashboard />,  label: '仪表盘' },
  { key: '/dns',                      icon: <IconCloud />,      label: 'DNS 管理',
    children: [
      { key: '/dns/accounts',    label: 'DNS 账户' },
      { key: '/dns/domains',     label: '域名列表' },
      { key: '/dns/records',     label: 'DNS 记录' },
      { key: '/dns/publish-jobs',label: '发布任务' },
    ],
  },
  { key: '/cdn',                      icon: <IconApps />,       label: 'CDN 管理',
    children: [
      { key: '/cdn/providers',   label: '服务商' },
      { key: '/cdn/sites',       label: 'CDN 站点' },
    ],
  },
  { key: '/scheduler',                icon: <IconCompass />,    label: '智能调度',
    children: [
      { key: '/scheduler/nodes',       label: 'CDN 节点' },
      { key: '/scheduler/node-pools',  label: '节点池' },
      { key: '/scheduler/probes',      label: '探针管理' },
      { key: '/scheduler/candidates',  label: 'IP 候选' },
    ],
  },
  { key: '/ssl',                      icon: <IconSafe />,       label: 'SSL 证书',
    children: [
      { key: '/ssl/certificates', label: '证书列表' },
      { key: '/ssl/workflows',    label: '工作流' },
    ],
  },
  { key: '/monitor',                  icon: <IconBulb />,       label: '监控中心',
    children: [
      { key: '/monitor/metrics',      label: '指标大盘' },
      { key: '/monitor/alert-rules',  label: '告警规则' },
      { key: '/monitor/alerts',       label: '告警管理' },
      { key: '/monitor/channels',     label: '通知渠道' },
    ],
  },
  { key: '/system',                   icon: <IconSettings />,   label: '系统管理',
    children: [
      { key: '/system/users',    label: '用户管理' },
      { key: '/system/audit',    label: '审计日志' },
    ],
  },
];

function findOpenKeys(pathname: string): string[] {
  for (const item of navItems) {
    if (item.children?.some((c) => pathname.startsWith(c.key))) {
      return [item.key];
    }
  }
  return [];
}

export default function MainLayout() {
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);
  const openKeys = findOpenKeys(location.pathname);

  return (
    <Layout style={{ height: '100vh' }}>
      <Sider
        collapsed={collapsed}
        onCollapse={setCollapsed}
        collapsible
        width={220}
        style={{ background: '#001529' }}
      >
        <div style={{ padding: collapsed ? '16px 8px' : '16px 24px', color: '#fff', fontWeight: 700, fontSize: 16 }}>
          {collapsed ? 'SC' : 'SmartCDN'}
        </div>
        <Menu
          theme="dark"
          selectedKeys={[location.pathname]}
          defaultOpenKeys={openKeys}
          style={{ width: '100%' }}
        >
          {navItems.map((item) =>
            item.children ? (
              <SubMenu key={item.key} title={<span>{item.icon} {!collapsed && item.label}</span>}>
                {item.children.map((child) => (
                  <MenuItem key={child.key}>
                    <Link to={child.key}>{child.label}</Link>
                  </MenuItem>
                ))}
              </SubMenu>
            ) : (
              <MenuItem key={item.key}>
                <Link to={item.key}>{item.icon} {!collapsed && item.label}</Link>
              </MenuItem>
            )
          )}
        </Menu>
      </Sider>

      <Layout>
        <Header style={{ background: '#fff', borderBottom: '1px solid #e8e8e8', padding: '0 24px', display: 'flex', alignItems: 'center' }}>
          <Typography.Title heading={6} style={{ margin: 0, flex: 1 }}>
            SmartCDN 管理控制台
          </Typography.Title>
        </Header>
        <Content style={{ padding: 24, overflow: 'auto', background: '#f5f5f5' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
