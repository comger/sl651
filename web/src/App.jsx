import React, { useState, useEffect } from 'react';
import { Layout, Table, Checkbox, Card, Typography, Space, Tag, theme, Divider, Empty, ConfigProvider, Statistic, Row, Col, Badge, Button, Modal, Form, InputNumber, Tooltip, Drawer, Timeline, Menu, Progress, Input, Select, message, Tabs, Descriptions, Popconfirm } from 'antd';
import { DatabaseOutlined, HddOutlined, SyncOutlined, SettingOutlined, DashboardOutlined, HistoryOutlined, AlertOutlined, SafetyCertificateOutlined, BugOutlined, ProfileOutlined, SignalFilled, ControlOutlined, SendOutlined, ReadOutlined, PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons';
import axios from 'axios';
import moment from 'moment';
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as ChartTooltip, Legend } from 'recharts';

const { Header, Content, Sider } = Layout;
const { Title, Text } = Typography;

// Configure axios base URL - Pointing to port 8085 as per config
const api = axios.create({
  baseURL: 'http://127.0.0.1:8085/api/v1',
});

const MonitoringCharts = ({ data, span = 12 }) => {
  if (!data || data.length === 0) return null;

  const keyMapping = {
    'water_level': '水位',
    'voltage': '电压',
    'cumulative_flow': '累计流量',
    'total_rainfall': '总雨量',
    'rainfall': '雨量',
    'instant_flow': '瞬时流量',
    'hourly_rainfall': '小时雨量',
    'day_rainfall': '日雨量',
    'period_rainfall': '时段雨量',
    'extra_rainfall': '额外雨量',
  };

  // Reverse data to show chronological order (left to right)
  const sortedData = [...data].reverse();

  // Pick some colors for multiple devices
  const colors = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#13c2c2', '#eb2f96'];

  // Extract all unique meaningful tags
  const allTags = new Set();
  sortedData.forEach(item => {
    if (Array.isArray(item.values)) {
      item.values.forEach(v => {
        const metadataTags = ['station_id', 'send_time', 'body_station_id', 'station_class', 'observation_time', 'serial', 'function_code', 'timestamp'];
        if (!metadataTags.includes(v.tag)) {
          allTags.add(v.tag);
        }
      });
    }
  });

  const deviceIds = Array.from(new Set(sortedData.map(d => d.device_id)));

  // Transform data for line charts: Array of { time, [deviceId1_tag]: val, [deviceId2_tag]: val, ... }
  // To keep it simple and handle timestamps correctly, we group by time
  const timeBuckets = {};
  sortedData.forEach(item => {
    const timeStr = moment(item.timestamp).format('HH:mm:ss');
    if (!timeBuckets[timeStr]) {
      timeBuckets[timeStr] = { time: timeStr };
    }
    if (Array.isArray(item.values)) {
      item.values.forEach(v => {
        const val = v.value.float ?? v.value.int;
        if (typeof val === 'number') {
          // Use deviceId as prefix for multi-line support on same tag
          timeBuckets[timeStr][`${item.device_id}_${v.tag}`] = val;
        }
      });
    }
  });

  const chartData = Object.values(timeBuckets).sort((a, b) => a.time.localeCompare(b.time));

  return (
    <Row gutter={[16, 16]} style={{ marginBottom: sortedData.length > 0 ? '20px' : '0' }}>
      {Array.from(allTags).map(tag => (
        <Col span={span} key={tag}>
          <Card title={`${keyMapping[tag] || tag} 实时趋势`} size="small" bordered={false} style={{ borderRadius: '8px' }}>
            <div style={{ width: '100%', height: 180 }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f0f0f0" />
                  <XAxis
                    dataKey="time"
                    fontSize={10}
                    tick={{ fill: '#8c8c8c' }}
                    axisLine={{ stroke: '#f0f0f0' }}
                  />
                  <YAxis
                    fontSize={10}
                    tick={{ fill: '#8c8c8c' }}
                    axisLine={{ stroke: '#f0f0f0' }}
                    domain={['auto', 'auto']}
                  />
                  <ChartTooltip
                    contentStyle={{ borderRadius: '4px', border: 'none', boxShadow: '0 2px 8px rgba(0,0,0,0.15)', fontSize: '11px' }}
                  />
                  <Legend verticalAlign="top" height={24} iconType="circle" wrapperStyle={{ fontSize: '10px' }} />
                  {deviceIds.map((devId, idx) => (
                    <Line
                      key={`${devId}_${tag}`}
                      type="monotone"
                      dataKey={`${devId}_${tag}`}
                      name={`设备 ${devId}`}
                      stroke={colors[idx % colors.length]}
                      strokeWidth={2}
                      dot={false}
                      activeDot={{ r: 4 }}
                      connectNulls={true}
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </Col>
      ))}
    </Row>
  );
};

const App = () => {
  const [devices, setDevices] = useState([]);
  const [selectedStationIDs, setSelectedStationIDs] = useState([]);
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [limit, setLimit] = useState(50);
  const [refreshInterval, setRefreshInterval] = useState(5);
  const [stats, setStats] = useState({ total: 0, online: 0, offline: 0 });
  const [hbStatus, setHbStatus] = useState({});
  const [hbConfig, setHbConfig] = useState({ check_interval: 60, default_timeout: 300 });
  const [faultLogs, setFaultLogs] = useState([]);
  const [systemLogs, setSystemLogs] = useState([]);
  const [qualityMetrics, setQualityMetrics] = useState([]);
  const [qualityStats, setQualityStats] = useState({});
  const [qualityDeviceStats, setQualityDeviceStats] = useState([]);
  const [qualityDuration, setQualityDuration] = useState('24h');
  const [isConfigModalOpen, setIsConfigModalOpen] = useState(false);
  const [isTrendDrawerOpen, setIsTrendDrawerOpen] = useState(false);
  const [trendData, setTrendData] = useState([]);
  const [activeTab, setActiveTab] = useState('monitoring');
  const [showLogs, setShowLogs] = useState(true);
  const [configLoading, setConfigLoading] = useState(false);
  const [controlHistory, setControlHistory] = useState([]);
  const [executingCmd, setExecutingCmd] = useState(false);
  const [controlDeviceID, setControlDeviceID] = useState('');
  const [forwardingRules, setForwardingRules] = useState([]);
  const [isRuleModalOpen, setIsRuleModalOpen] = useState(false);
  const [editingRule, setEditingRule] = useState(null);
  const [forwardingLogs, setForwardingLogs] = useState([]);
  const [isLogDetailOpen, setIsLogDetailOpen] = useState(false);
  const [selectedLog, setSelectedLog] = useState(null);
  const [form] = Form.useForm();
  const [cmdForm] = Form.useForm();
  const [ruleForm] = Form.useForm();

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // Fetch metadata and stats
  useEffect(() => {
    fetchDevices();
    fetchStats();
    fetchHbStatus();
    fetchHbConfig();
    const interval = setInterval(() => {
      fetchDevices();
      fetchStats();
      fetchHbStatus();
      if (activeTab === 'diagnosis') {
        fetchFaultLogs();
        fetchSystemLogs();
      }
      if (activeTab === 'quality') {
        fetchQualityMetrics();
        fetchQualityStats();
        fetchQualityDeviceStats();
      }
      if (activeTab === 'control') {
        fetchControlHistory(controlDeviceID);
      }
      if (activeTab === 'forwarding') {
        fetchForwardingRules();
      }
    }, 10000);
    return () => clearInterval(interval);
  }, [activeTab, controlDeviceID]);

  // Poll data based on selection
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, refreshInterval * 1000);
    return () => clearInterval(interval);
  }, [selectedStationIDs, limit, refreshInterval]);

  const fetchDevices = async () => {
    try {
      const res = await api.get('/devices');
      if (res.data.code === 0) {
        const devList = (res.data.data || []).sort((a, b) => a.id.localeCompare(b.id));
        setDevices(devList);
        // Auto-select all if selection is empty (first load)
        if (selectedStationIDs.length === 0 && devList.length > 0) {
          setSelectedStationIDs(devList.map(d => d.id));
        }
      }
    } catch (error) {
      console.error('Failed to fetch devices:', error);
    }
  };

  const fetchStats = async () => {
    try {
      const res = await api.get('/statistics/online');
      if (res.data.code === 0) {
        setStats(res.data.data);
      }
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  const fetchHbStatus = async () => {
    try {
      const res = await api.get('/heartbeat/status');
      if (res.data.code === 0) {
        setHbStatus(res.data.data || {});
      }
    } catch (error) {
      console.error('Failed to fetch hb status:', error);
    }
  };

  const fetchHbConfig = async () => {
    try {
      const res = await api.get('/heartbeat/config');
      if (res.data.code === 0) {
        setHbConfig(res.data.data);
        form.setFieldsValue(res.data.data);
      }
    } catch (error) {
      console.error('Failed to fetch hb config:', error);
    }
  };

  const updateHbConfig = async (values) => {
    setConfigLoading(true);
    try {
      const res = await api.put('/heartbeat/config', values);
      if (res.data.code === 0) {
        setHbConfig(values);
        setIsConfigModalOpen(false);
      }
    } catch (error) {
      console.error('Failed to update hb config:', error);
    } finally {
      setConfigLoading(false);
    }
  };

  const fetchTrendData = async () => {
    try {
      const res = await api.get('/statistics/trend');
      if (res.data.code === 0) {
        setTrendData(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch trend:', error);
    }
  };

  const fetchFaultLogs = async () => {
    try {
      const res = await api.get('/diagnosis/faults', { params: { limit: 100 } });
      if (res.data.code === 0) {
        setFaultLogs(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch fault logs:', error);
    }
  };

  const fetchSystemLogs = async () => {
    try {
      const res = await api.get('/diagnosis/logs', { params: { limit: 100 } });
      if (res.data.code === 0) {
        setSystemLogs(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch system logs:', error);
    }
  };

  const fetchQualityMetrics = async () => {
    try {
      const params = { limit: 100 };
      if (selectedStationIDs.length === 1) {
        params.device_id = selectedStationIDs[0];
      }
      const res = await api.get('/diagnosis/quality', { params });
      if (res.data.code === 0) {
        setQualityMetrics(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch quality metrics:', error);
    }
  };

  const fetchQualityStats = async () => {
    try {
      const params = { duration: qualityDuration };
      if (selectedStationIDs.length === 1) {
        params.device_id = selectedStationIDs[0];
      }
      const res = await api.get('/diagnosis/quality/stats', { params });
      if (res.data.code === 0) {
        setQualityStats(res.data.data || {});
      }
      // Also update the device-level list whenever analysis is triggered
      fetchQualityDeviceStats();
    } catch (error) {
      console.error('Failed to fetch quality stats:', error);
    }
  };

  const fetchQualityDeviceStats = async () => {
    try {
      const params = { duration: qualityDuration };
      const res = await api.get('/diagnosis/quality/devices', { params });
      if (res.data.code === 0) {
        setQualityDeviceStats(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch quality device stats:', error);
    }
  };

  const fetchControlHistory = async (deviceID = '') => {
    try {
      const params = deviceID ? { device_id: deviceID } : {};
      const res = await api.get('/diagnosis/control/history', { params });
      if (res.data.code === 0) {
        setControlHistory(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch control history:', error);
    }
  };

  const handleSendCommand = async (values) => {
    setExecutingCmd(true);
    try {
      const res = await api.post('/diagnosis/control/command', {
        device_id: values.device_id,
        function_code: values.function_code,
        payload: values.payload
      });
      if (res.data.code === 0) {
        message.success('指令已加入队列，等待设备上报时下发');
        fetchControlHistory(values.device_id);
      }
    } catch (error) {
      console.error('Failed to send command:', error);
      message.error('发送指令失败');
    } finally {
      setExecutingCmd(false);
    }
  };

  const fetchForwardingRules = async () => {
    try {
      const res = await api.get('/forwarding/rules');
      if (res.data.code === 0) {
        setForwardingRules(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch forwarding rules:', error);
    }
  };

  const fetchForwardingLogs = async () => {
    try {
      const res = await api.get('/forwarding/logs');
      if (res.data.code === 0) {
        setForwardingLogs(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch forwarding logs:', error);
    }
  };

  const handleSaveForwardingRule = async (values) => {
    try {
      const ruleData = {
        ...values,
        enabled: values.enabled !== false,
      };

      let res;
      if (editingRule) {
        res = await api.put(`/forwarding/rules/${editingRule.id}`, ruleData);
      } else {
        res = await api.post('/forwarding/rules', ruleData);
      }

      if (res.data.code === 0) {
        message.success('转发规则已保存');
        setIsRuleModalOpen(false);
        fetchForwardingRules();
      }
    } catch (error) {
      console.error('Failed to save forwarding rule:', error);
      message.error('保存失败');
    }
  };

  const handleDeleteForwardingRule = async (id) => {
    try {
      const res = await api.delete(`/forwarding/rules/${id}`);
      if (res.data.code === 0) {
        message.success('规则已删除');
        fetchForwardingRules();
      }
    } catch (error) {
      console.error('Failed to delete forwarding rule:', error);
      message.error('删除失败');
    }
  };

  const getDSNExample = (driver) => {
    switch (driver) {
      case 'mysql': return 'mysql://user:password@localhost:3306/dbname?parseTime=true';
      case 'postgres': return 'postgres://user:password@localhost:5432/dbname?sslmode=disable';
      case 'sqlite': return 'data/platform.db';
      default: return null;
    }
  };

  const openTrendDrawer = async () => {
    await fetchTrendData();
    setIsTrendDrawerOpen(true);
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      if (selectedStationIDs.length === 0) {
        // No need to fetch if empty, but we can set data to empty to be safe
        setData([]);
        setLoading(false);
        return;
      }
      const params = { limit: limit };
      if (selectedStationIDs.length > 0) {
        params.station_ids = selectedStationIDs.join(',');
      }
      const res = await api.get('/data', { params });
      if (res.data.code === 0) {
        const rawData = res.data.data || [];
        // Ensure values is parsed if it's a string
        const processedData = rawData.map(item => {
          if (typeof item.values === 'string') {
            try {
              item.values = JSON.parse(item.values);
            } catch (e) {
              item.values = [];
            }
          }
          return item;
        });
        setData(processedData);
      }
    } catch (error) {
      console.error('Failed to fetch data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDeviceChange = (checkedValues) => {
    setSelectedStationIDs(checkedValues);
  };

  const columns = [
    {
      title: '报文数据监测 (原始 \u0026 解析)',
      key: 'content',
      render: (_, record) => {
        const rawText = record.raw_data || '-';
        const values = record.values || [];

        // Extract metadata for the sub-line
        const sendTimeObj = values.find(v => v.tag === 'send_time');
        const sendTime = sendTimeObj && sendTimeObj.value && sendTimeObj.value.string ? sendTimeObj.value.string : '-';
        const storeTime = moment(record.timestamp).format('MM-DD HH:mm:ss');

        const keyMapping = {
          'water_level': '水位',
          'voltage': '电压',
          'cumulative_flow': '累计流量',
          'total_rainfall': '总雨量',
          'rainfall': '雨量',
          'instant_flow': '瞬时流量',
          'hourly_rainfall': '小时雨量',
          'day_rainfall': '日雨量',
          'period_rainfall': '时段雨量',
          'extra_rainfall': '额外雨量',
        };

        const parsedTags = (
          <Space size={[4, 4]} wrap style={{ marginTop: 4 }}>
            {values.map((kp, index) => {
              const label = keyMapping[kp.tag] || kp.tag;
              if (['station_id', 'send_time', 'body_station_id', 'station_class', 'observation_time', 'serial', 'function_code', 'timestamp'].includes(kp.tag)) return null;
              if (!kp.value) return null;

              let val = kp.value.float ?? kp.value.int ?? kp.value.string ?? kp.value.bool;
              if (typeof val === 'number' && !Number.isInteger(val)) {
                val = val.toFixed(3);
              }
              return (
                <Tag key={index} color="processing" style={{ margin: 0, fontSize: '11px', borderRadius: '2px' }}>
                  {label}: {val}
                </Tag>
              );
            })}
          </Space>
        );

        return (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '2px', padding: '4px 0' }}>
            <Text code copyable style={{ fontSize: '12px', whiteSpace: 'pre-wrap', wordBreak: 'break-all', color: '#10239e' }}>
              {rawText}
            </Text>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '11px', color: '#8c8c8c' }}>
              <Tag size="small" color="blue" style={{ margin: 0, fontSize: '10px' }}>{record.device_id}</Tag>
              <Tag size="small" color={record.direction === 'downlink' ? 'orange' : 'green'} style={{ margin: 0, fontSize: '10px' }}>
                {record.direction === 'downlink' ? '下行' : '上行'}
              </Tag>
              {record.function_code && (
                <Tag size="small" color="purple" style={{ margin: 0, fontSize: '10px' }}>
                  FC: {record.function_code}
                </Tag>
              )}
              <span>采集: {sendTime}</span>
              <Divider type="vertical" style={{ margin: 0 }} />
              <span>入库: {storeTime}</span>
            </div>
            {parsedTags}
          </div>
        );
      },
    },
  ];

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1890ff',
        },
      }}
    >
      <Layout style={{ minHeight: '100vh' }}>
        <Header style={{ display: 'flex', alignItems: 'center', background: '#001529', padding: '0 20px', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <DatabaseOutlined style={{ color: '#fff', fontSize: '24px', marginRight: '10px' }} />
            <Title level={4} style={{ color: '#fff', margin: 0, marginRight: '40px' }}>SL651 监测与解析平台</Title>
            <Menu
              theme="dark"
              mode="horizontal"
              defaultSelectedKeys={['monitoring']}
              selectedKeys={[activeTab]}
              onClick={(e) => setActiveTab(e.key)}
              style={{ minWidth: '300px' }}
              items={[
                { key: 'monitoring', icon: <DashboardOutlined />, label: '实时监测' },
                { key: 'forwarding', icon: <SendOutlined />, label: '转发配置' },
                { key: 'diagnosis', icon: <BugOutlined />, label: '运维审计' },
                { key: 'quality', icon: <SignalFilled />, label: '质量评估' },
                { key: 'control', icon: <ControlOutlined />, label: '远程配置' },
              ]}
            />
          </div>
          <Space>
            <Tooltip title="当前系统健康状况">
              <Badge count={faultLogs.filter(f => f.severity === 'error' || f.severity === 'critical').length} offset={[10, 0]}>
                <SafetyCertificateOutlined style={{ color: '#52c41a', fontSize: '20px' }} />
              </Badge>
            </Tooltip>
          </Space>
        </Header>
        <Layout hasSider>
          <Sider width={280} style={{ background: colorBgContainer, padding: '20px', borderRight: '1px solid #f0f0f0', overflowY: 'auto' }}>
            <div style={{ marginBottom: '20px', display: 'flex', alignItems: 'center' }}>
              <HddOutlined style={{ marginRight: '8px', fontSize: '18px' }} />
              <Title level={4} style={{ margin: 0 }}>设备分组管家</Title>
            </div>
            <Divider style={{ margin: '10px 0' }} orientation="left">在线设备列表</Divider>
            {devices.length > 0 ? (
              <Checkbox.Group
                style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}
                value={selectedStationIDs}
                onChange={handleDeviceChange}
              >
                {devices.map(dev => {
                  const status = hbStatus[dev.id];
                  const isOnline = status ? status.online : (dev.status === 'online');
                  return (
                    <Card
                      key={dev.id}
                      size="small"
                      hoverable
                      style={{
                        borderLeft: `4px solid ${isOnline ? '#52c41a' : '#bfbfbf'}`,
                        marginBottom: '8px'
                      }}
                      bodyStyle={{ padding: '8px' }}
                    >
                      <Checkbox value={dev.id} style={{ width: '100%', marginLeft: 0 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', width: '100%' }}>
                          <Space direction="vertical" size={0}>
                            <Text strong>{dev.name || '未命名设备'}</Text>
                            <Text type="secondary" style={{ fontSize: '11px' }}>ID: {dev.id}</Text>
                            <Text type="secondary" style={{ fontSize: '10px' }}>
                              连接: {status && status.last_seen ? moment(status.last_seen).format('MM-DD HH:mm:ss') : '-'}
                            </Text>
                          </Space>
                          <Badge
                            status={isOnline ? 'processing' : 'default'}
                            text={<Text style={{ fontSize: '11px' }} type={isOnline ? 'success' : 'secondary'}>{isOnline ? '在线' : '离线'}</Text>}
                          />
                        </div>
                      </Checkbox>
                    </Card>
                  );
                })}
              </Checkbox.Group>
            ) : (
              <Empty description="暂无在线设备" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Sider>
          <Content style={{ padding: '24px', overflowY: 'auto', background: '#f0f2f5', display: 'flex', flexDirection: 'column', gap: '20px' }}>
            {activeTab === 'monitoring' && (
              <>
                <Row gutter={16}>
                  <Col span={8}>
                    <Card bordered={false} style={{ borderRadius: borderRadiusLG }}>
                      <Statistic
                        title="总设备数"
                        value={stats.total}
                        prefix={<HddOutlined />}
                        valueStyle={{ color: '#001529' }}
                      />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card bordered={false} style={{ borderRadius: borderRadiusLG }}>
                      <Statistic
                        title="当前在线"
                        value={stats.online}
                        prefix={<SyncOutlined spin={stats.online > 0} />}
                        valueStyle={{ color: '#52c41a' }}
                      />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card bordered={false} style={{ borderRadius: borderRadiusLG }}>
                      <Statistic
                        title="离线警告"
                        value={stats.offline}
                        prefix={<HistoryOutlined />}
                        valueStyle={{ color: stats.offline > 0 ? '#ff4d4f' : '#8c8c8c' }}
                      />
                    </Card>
                  </Col>
                </Row>

                <Row gutter={[16, 16]} style={{ flex: 1, overflow: 'hidden' }}>
                  <Col span={showLogs ? 14 : 24} style={{ height: '100%', overflowY: 'auto', paddingRight: showLogs ? '8px' : '0' }}>
                    <MonitoringCharts data={data} span={showLogs ? 24 : 12} />
                  </Col>

                  {showLogs && (
                    <Col span={10} style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
                      <Card
                        bordered={false}
                        style={{ borderRadius: borderRadiusLG, boxShadow: '0 1px 2px 0 rgba(0,0,0,0.03)', flex: 1, display: 'flex', flexDirection: 'column' }}
                        bodyStyle={{ padding: '0', flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}
                      >
                        <div style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <Space>
                            <SyncOutlined spin={loading} />
                            <Title level={5} style={{ margin: 0, fontSize: '14px' }}>实时报文日志</Title>
                          </Space>
                          <Space>
                            <Tooltip title="隐藏日志">
                              <Button type="text" size="small" icon={<ProfileOutlined />} onClick={() => setShowLogs(false)} />
                            </Tooltip>
                          </Space>
                        </div>
                        <Table
                          columns={columns}
                          dataSource={data}
                          rowKey="id"
                          pagination={false}
                          size="small"
                          style={{ flex: 1 }}
                          scroll={{ y: 'calc(100vh - 380px)' }}
                        />
                      </Card>
                    </Col>
                  )}
                </Row>

                {!showLogs && (
                  <div style={{ position: 'fixed', right: '40px', bottom: '40px', zIndex: 100 }}>
                    <Button
                      type="primary"
                      shape="circle"
                      size="large"
                      icon={<ProfileOutlined />}
                      onClick={() => setShowLogs(true)}
                      style={{ boxShadow: '0 4px 12px rgba(0,0,0,0.15)' }}
                    />
                  </div>
                )}
              </>
            )}
            {activeTab === 'forwarding' && (
              <Card bordered={false} style={{ borderRadius: borderRadiusLG }}>
                <Tabs
                  defaultActiveKey="rules"
                  onChange={(key) => {
                    if (key === 'logs') fetchForwardingLogs();
                    else fetchForwardingRules();
                  }}
                >
                  <Tabs.TabPane tab="规则管理" key="rules">
                    <div style={{ padding: '0 0 16px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Title level={4} style={{ margin: 0 }}>转发配置与北向中心</Title>
                      <Button type="primary" icon={<PlusOutlined />} onClick={() => {
                        setEditingRule(null);
                        ruleForm.resetFields();
                        ruleForm.setFieldsValue({
                          destinations: [{ dest_type: 'mqtt', url: 'tcp://broker.hivemq.com:1883' }],
                          enabled: true
                        });
                        setIsRuleModalOpen(true);
                      }}>
                        新增转发规则
                      </Button>
                    </div>
                    <Table
                      dataSource={forwardingRules}
                      rowKey="id"
                      bordered={false}
                      columns={[
                        { title: '规则名称', dataIndex: 'name', key: 'name', render: (n, r) => <Space><Text strong>{n}</Text>{!r.enabled && <Tag>已禁用</Tag>}</Space> },
                        {
                          title: '转发目标',
                          dataIndex: 'destinations',
                          key: 'destinations',
                          render: (dests) => (
                            <Space direction="vertical" size={2}>
                              {(dests || []).map((d, i) => (
                                <Tag key={i} color={['mysql', 'postgres', 'sqlite', 'database'].includes(d.dest_type) ? 'blue' : d.dest_type === 'mqtt' ? 'green' : 'orange'}>
                                  {d.dest_type.toUpperCase()}: {d.url}
                                </Tag>
                              ))}
                            </Space>
                          )
                        },
                        {
                          title: '状态',
                          dataIndex: 'enabled',
                          key: 'enabled',
                          render: (e) => <Badge status={e ? 'success' : 'default'} text={e ? '已启用' : '已停用'} />
                        },
                        {
                          title: '操作',
                          key: 'action',
                          width: 150,
                          render: (_, record) => (
                            <Space>
                              <Button type="text" icon={<EditOutlined />} onClick={() => {
                                setEditingRule(record);
                                ruleForm.setFieldsValue(record);
                                setIsRuleModalOpen(true);
                              }} />
                              <Popconfirm
                                title="确认删除规则?"
                                description={`规则 "${record.name}" 将被永久移除。`}
                                onConfirm={() => handleDeleteForwardingRule(record.id)}
                              >
                                <Button type="text" danger icon={<DeleteOutlined />} />
                              </Popconfirm>
                            </Space>
                          )
                        }
                      ]}
                    />
                  </Tabs.TabPane>
                  <Tabs.TabPane tab="转发日志" key="logs">
                    <div style={{ padding: '0 0 16px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Title level={4} style={{ margin: 0 }}>转发执行动态</Title>
                      <Button icon={<SyncOutlined />} onClick={fetchForwardingLogs}>刷新日志</Button>
                    </div>
                    <Table
                      dataSource={forwardingLogs}
                      rowKey="id"
                      pagination={{ pageSize: 12 }}
                      columns={[
                        { title: '时间', dataIndex: 'created_at', key: 'time', render: (t) => moment(t).format('YYYY-MM-DD HH:mm:ss') },
                        { title: '状态', dataIndex: 'status', key: 'status', render: (s) => <Tag color={s === 'success' ? 'success' : 'error'}>{s === 'success' ? '成功' : '失败'}</Tag> },
                        { title: '目标类型', dataIndex: 'target_type', key: 'target_type', render: (t) => <Tag>{t?.toUpperCase()}</Tag> },
                        { title: '目的地', dataIndex: 'destination_url', key: 'url', ellipsis: true },
                        {
                          title: '操作',
                          key: 'action',
                          render: (_, record) => (
                            <Button type="link" size="small" onClick={() => {
                              setSelectedLog(record);
                              setIsLogDetailOpen(true);
                            }}>查看详情</Button>
                          )
                        },
                      ]}
                    />
                  </Tabs.TabPane>
                </Tabs>
              </Card>
            )}
            {activeTab === 'diagnosis' && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                <Row gutter={16}>
                  <Col span={12}>
                    <Card title={<span><AlertOutlined style={{ color: '#ff4d4f' }} /> 设备故障日志</span>} bordered={false} style={{ borderRadius: borderRadiusLG }}>
                      <Table
                        dataSource={faultLogs}
                        rowKey="id"
                        size="small"
                        pagination={{ pageSize: 10 }}
                        columns={[
                          { title: '故障代码', dataIndex: 'fault_code', width: 110, render: code => <Text code>{code || '-'}</Text> },
                          { title: '时间', dataIndex: 'time', width: 140, render: t => moment(t).format('MM-DD HH:mm:ss') },
                          { title: '设备ID', dataIndex: 'device_id', width: 100 },
                          {
                            title: '级别', dataIndex: 'severity', width: 80, render: s => (
                              <Tag color={s === 'critical' ? 'volcano' : s === 'error' ? 'red' : 'orange'}>{s?.toUpperCase()}</Tag>
                            )
                          },
                          { title: '故障描述', dataIndex: 'message' },
                        ]}
                      />
                    </Card>
                  </Col>
                  <Col span={12}>
                    <Card title={<span><ProfileOutlined /> 系统运行日志</span>} bordered={false} style={{ borderRadius: borderRadiusLG }}>
                      <Table
                        dataSource={systemLogs}
                        rowKey="id"
                        size="small"
                        pagination={{ pageSize: 10 }}
                        columns={[
                          { title: '时间', dataIndex: 'time', width: 140, render: t => moment(t).format('MM-DD HH:mm:ss') },
                          { title: '模块', dataIndex: 'source', width: 100 },
                          { title: '级别', dataIndex: 'level', width: 80, render: l => <Tag>{l}</Tag> },
                          { title: '内容', dataIndex: 'message' },
                        ]}
                      />
                    </Card>
                  </Col>
                </Row>
              </div>
            )}
            {activeTab === 'quality' && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                <Row gutter={16}>
                  <Col span={24}>
                    <Card
                      title={<span><SignalFilled style={{ color: '#1890ff' }} /> 报文质量统计分析</span>}
                      extra={
                        <Space>
                          <Text type="secondary">分析周期:</Text>
                          <select
                            value={qualityDuration}
                            onChange={(e) => setQualityDuration(e.target.value)}
                            style={{ padding: '2px 8px', borderRadius: '4px', border: '1px solid #d9d9d9' }}
                          >
                            <option value="1h">最近1小时</option>
                            <option value="6h">最近6小时</option>
                            <option value="24h">最近24小时</option>
                            <option value="168h">最近7天</option>
                          </select>
                          <Button size="small" type="primary" onClick={fetchQualityStats} ghost icon={<SyncOutlined />}>立即分析</Button>
                        </Space>
                      }
                      bordered={false}
                      style={{ borderRadius: borderRadiusLG }}
                    >
                      <Row gutter={16}>
                        <Col span={4}>
                          <Statistic title="综合质量评分" value={qualityStats.avg_score || 0} precision={2} suffix="/ 100" valueStyle={{ color: (qualityStats.avg_score || 0) > 80 ? '#3f8600' : '#cf1322' }} />
                        </Col>
                        <Col span={4}>
                          <Statistic title="平均完整度" value={qualityStats.avg_completeness || 0} precision={2} suffix="%" />
                        </Col>
                        <Col span={4}>
                          <Statistic title="平均延时" value={qualityStats.avg_latency || 0} precision={3} suffix="s" />
                        </Col>
                        <Col span={4}>
                          <Statistic title="平均误码率" value={qualityStats.avg_error_rate || 0} precision={3} suffix="%" valueStyle={{ color: (qualityStats.avg_error_rate || 0) < 1 ? '#3f8600' : '#cf1322' }} />
                        </Col>
                        <Col span={4}>
                          <Statistic title="平均抖动" value={qualityStats.avg_jitter || 0} precision={3} suffix="s" />
                        </Col>
                        <Col span={4}>
                          <Statistic title="样本总数" value={qualityStats.count || 0} prefix={<DatabaseOutlined />} />
                        </Col>
                      </Row>
                    </Card>
                  </Col>
                </Row>

                <Card title={<span><HistoryOutlined /> 各设备质量评估 (按周期统计)</span>} bordered={false} style={{ borderRadius: borderRadiusLG }}>
                  <Table
                    dataSource={qualityDeviceStats}
                    rowKey="device_id"
                    pagination={{ pageSize: 10 }}
                    size="middle"
                    columns={[
                      { title: '设备ID', dataIndex: 'device_id', width: 120, render: id => <Tag color="blue">{id}</Tag> },
                      { title: '综合评分', dataIndex: 'avg_score', width: 150, render: s => <Progress percent={Math.round(s)} size="small" status={s < 60 ? 'exception' : s < 85 ? 'normal' : 'success'} /> },
                      { title: '平均完整度', dataIndex: 'avg_completeness', width: 110, render: c => <Tag color={c > 99 ? 'green' : 'orange'}>{c.toFixed(2)}%</Tag> },
                      { title: '平均延时', dataIndex: 'avg_latency', width: 100, render: l => <Tag color={l < 3 ? 'cyan' : 'red'}>{l.toFixed(3)}s</Tag> },
                      { title: '误码率', dataIndex: 'avg_error_rate', width: 100, render: e => <Text type={e > 1 ? 'danger' : ''}>{e.toFixed(3)}%</Text> },
                      { title: '抖动', dataIndex: 'avg_jitter', width: 100, render: j => <Text secondary>{j.toFixed(4)}s</Text> },
                      { title: '报文总数', dataIndex: 'count', width: 100 },
                      { title: '最后评估', dataIndex: 'last_time', render: t => moment(t).format('HH:mm:ss') },
                    ]}
                  />
                </Card>
              </div>
            )}
            {activeTab === 'control' && (
              <div style={{ padding: '0 20px', display: 'flex', flexDirection: 'column', gap: '20px' }}>
                <Row gutter={20}>
                  <Col span={8}>
                    <Card title={<span><ControlOutlined /> 下发控制指令</span>} bordered={false} style={{ borderRadius: borderRadiusLG }}>
                      <Form form={cmdForm} layout="vertical" onFinish={handleSendCommand}>
                        <Form.Item name="device_id" label="目标设备" rules={[{ required: true }]}>
                          <Select
                            placeholder="请选择目标设备"
                            options={devices.map(d => ({ label: `${d.name} (${d.id})`, value: d.id }))}
                            onChange={v => {
                              setControlDeviceID(v);
                              fetchControlHistory(v);
                            }}
                          />
                        </Form.Item>
                        <Form.Item name="function_code" label="功能码" rules={[{ required: true }]}>
                          <Select placeholder="请选择指令类型">
                            <Select.Option value="40">读取工作参数 (40H)</Select.Option>
                            <Select.Option value="41">读取运行状态 (41H)</Select.Option>
                            <Select.Option value="42">修改工作参数 (42H)</Select.Option>
                            <Select.Option value="43">修改运行状态 (43H)</Select.Option>
                          </Select>
                        </Form.Item>
                        <Form.Item name="payload" label="参数内容 (HEX)" help="请输入符合SL651规约的TLV数据或参数标识符">
                          <Input.TextArea rows={4} placeholder="例如: 010203 (读取配置1,2,3)" />
                        </Form.Item>
                        <Button type="primary" htmlType="submit" icon={<SendOutlined />} loading={executingCmd} block>
                          加入发送队列
                        </Button>
                      </Form>
                    </Card>
                    <Card title="常用指令快速配置" style={{ marginTop: '20px', borderRadius: borderRadiusLG }}>
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <Button size="small" onClick={() => cmdForm.setFieldsValue({ function_code: '40', payload: '01020304' })}>读取站号与主中心 (01-04)</Button>
                        <Button size="small" onClick={() => cmdForm.setFieldsValue({ function_code: '40', payload: '14' })}>读取定时报间隔 (14H)</Button>
                      </Space>
                    </Card>
                  </Col>
                  <Col span={16}>
                    <Card title={<span><HistoryOutlined /> 指令下发历史</span>} bordered={false} style={{ borderRadius: borderRadiusLG }}>
                      <Table
                        dataSource={controlHistory}
                        rowKey="id"
                        size="small"
                        pagination={{ pageSize: 12 }}
                        columns={[
                          { title: '指令ID', dataIndex: 'id', width: 140, render: id => <Text type="secondary" style={{ fontSize: '10px' }}>{id}</Text> },
                          { title: '功能', dataIndex: 'function_code', width: 100, render: c => <Tag color="blue">{c}H</Tag> },
                          {
                            title: '状态', dataIndex: 'status', width: 100, render: s => (
                              <Tag color={s === 'success' ? 'green' : s === 'sent' ? 'processing' : s === 'pending' ? 'orange' : 'red'}>
                                {s === 'success' ? '成功' : s === 'sent' ? '已下发' : s === 'pending' ? '待机' : '失败'}
                              </Tag>
                            )
                          },
                          { title: '载荷', dataIndex: 'payload', ellipsis: true },
                          { title: '结果', dataIndex: 'result', ellipsis: true, render: r => r || '-' },
                          { title: '创建时间', dataIndex: 'created_at', render: t => moment(t).format('HH:mm:ss') },
                        ]}
                      />
                    </Card>
                  </Col>
                </Row>
              </div>
            )}


            <Modal
              title={<span><SettingOutlined /> 心跳监测全球配置</span>}
              open={isConfigModalOpen}
              onOk={() => form.submit()}
              onCancel={() => setIsConfigModalOpen(false)}
              confirmLoading={configLoading}
              okText="保存配置"
              cancelText="取消"
            >
              <Form
                form={form}
                layout="vertical"
                onFinish={updateHbConfig}
                initialValues={hbConfig}
              >
                <Form.Item
                  name="check_interval"
                  label="监测巡检间隔 (秒)"
                  extra="系统对所有设备进行活跃度盘点的时间间隔"
                  rules={[{ required: true, message: '请输入巡检间隔' }]}
                >
                  <InputNumber min={5} max={3600} style={{ width: '100%' }} />
                </Form.Item>
                <Form.Item
                  name="default_timeout"
                  label="默认超时判定 (秒)"
                  extra="超过此时间未收到任何报文即判定为离线"
                  rules={[{ required: true, message: '请输入判定超时' }]}
                >
                  <InputNumber min={10} max={86400} style={{ width: '100%' }} />
                </Form.Item>
              </Form>
            </Modal>
            <Drawer
              title={<span><HistoryOutlined /> 设备运行轨迹 (最近24小时)</span>}
              placement="right"
              onClose={() => setIsTrendDrawerOpen(false)}
              open={isTrendDrawerOpen}
              width={400}
            >
              {trendData.length > 0 ? (
                <Timeline
                  items={trendData.map((item) => ({
                    color: item.status === 'online' ? 'green' : 'gray',
                    children: (
                      <div>
                        <Text strong>[{item.device_id}]</Text> 变为 <Tag color={item.status === 'online' ? 'success' : 'default'}>{item.status === 'online' ? '在线' : '离线'}</Tag>
                        <br />
                        <Text type="secondary" style={{ fontSize: '12px' }}>{moment(item.time).format('MM-DD HH:mm:ss')}</Text>
                      </div>
                    ),
                  })).reverse()}
                />
              ) : (
                <Empty description="暂无历史轨迹记录" />
              )}
            </Drawer>

            <Modal
              title={editingRule ? '编辑转发规则' : '新增转发规则'}
              open={isRuleModalOpen}
              onOk={() => ruleForm.submit()}
              onCancel={() => setIsRuleModalOpen(false)}
              width={700}
              okText="确认保存"
              cancelText="取消"
            >
              <Form form={ruleForm} layout="vertical" onFinish={handleSaveForwardingRule}>
                <Row gutter={16}>
                  <Col span={16}>
                    <Form.Item name="name" label="规则名称" rules={[{ required: true }]}>
                      <Input placeholder="输入转发规则名称, e.g. 数据转发至省中心" />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item name="enabled" label="是否启用" valuePropName="checked">
                      <Checkbox>启用该规则</Checkbox>
                    </Form.Item>
                  </Col>
                </Row>

                <Divider orientation="left" style={{ margin: '12px 0' }}>转发目标配置</Divider>
                <Form.List name="destinations">
                  {(fields, { add, remove }) => (
                    <>
                      {fields.map(({ key, name, ...restField }) => (
                        <Card size="small" key={key} style={{ marginBottom: 12, background: '#fafafa' }}
                          extra={<DeleteOutlined onClick={() => remove(name)} style={{ color: '#ff4d4f' }} />}>
                          <Row gutter={12}>
                            <Col span={6}>
                              <Form.Item
                                {...restField}
                                name={[name, 'dest_type']}
                                label="目标类型"
                                rules={[{ required: true }]}
                              >
                                <Select>
                                  <Select.Option value="mqtt">MQTT 代理</Select.Option>
                                  <Select.Option value="mysql">MySQL 数据库</Select.Option>
                                  <Select.Option value="postgres">PostgreSQL 数据库</Select.Option>
                                  <Select.Option value="sqlite">SQLite 数据库</Select.Option>
                                  <Select.Option value="http">HTTP Webhook</Select.Option>
                                </Select>
                              </Form.Item>
                            </Col>
                            <Form.Item noStyle shouldUpdate={(prevValues, currentValues) => prevValues.destinations !== currentValues.destinations}>
                              {({ getFieldValue }) => {
                                const destType = getFieldValue(['destinations', name, 'dest_type']);
                                const isDatabase = ['mysql', 'postgres', 'sqlite', 'database'].includes(destType);
                                if (isDatabase) {
                                  return (
                                    <Col span={18}>
                                      <Form.Item
                                        {...restField}
                                        name={[name, 'url']}
                                        label="连接地址"
                                        extra={(() => {
                                          const example = getDSNExample(destType === 'database' ? 'mysql' : destType);
                                          return example ? (
                                            <span>{`示例: `}<Text code copyable>{example}</Text></span>
                                          ) : null;
                                        })()}
                                        rules={[{ required: true }]}
                                      >
                                        <Input placeholder="输入 DSN 连接字符串" />
                                      </Form.Item>
                                    </Col>
                                  );
                                }
                                return (
                                  <Col span={18}>
                                    <Form.Item
                                      {...restField}
                                      name={[name, 'url']}
                                      label="连接地址 (URL/Host)"
                                      rules={[{ required: true }]}
                                    >
                                      <Input placeholder="tcp://host:1883 or http://api.example.com" />
                                    </Form.Item>
                                  </Col>
                                );
                              }}
                            </Form.Item>
                          </Row>
                        </Card>
                      ))}
                      <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                        添加转发目的地 (多中心转发)
                      </Button>
                    </>
                  )}
                </Form.List>

                <Divider orientation="left" style={{ margin: '12px 0' }}>过滤条件 (可选)</Divider>
                <Row gutter={12}>
                  <Col span={12}>
                    <Form.Item name={['filter', 'device_ids']} label="限制设备 ID (逗号分隔)">
                      <Input placeholder="留空转发所有设备" />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item name={['filter', 'data_types']} label="数据类型过滤">
                      <Select mode="multiple" placeholder="默认全部">
                        <Select.Option value="realtime">实时报</Select.Option>
                        <Select.Option value="alarm">报警报</Select.Option>
                        <Select.Option value="status">状态报</Select.Option>
                      </Select>
                    </Form.Item>
                  </Col>
                </Row>
              </Form>
            </Modal>

            <Modal
              title="转发日志详情"
              open={isLogDetailOpen}
              onCancel={() => setIsLogDetailOpen(false)}
              footer={[
                <Button key="close" onClick={() => setIsLogDetailOpen(false)}>关闭</Button>
              ]}
              width={800}
            >
              {selectedLog && (
                <Descriptions bordered column={1} size="small">
                  <Descriptions.Item label="创建时间">
                    {moment(selectedLog.created_at).format('YYYY-MM-DD HH:mm:ss')}
                  </Descriptions.Item>
                  <Descriptions.Item label="目标中心">
                    <Tag color="blue">{selectedLog.target_type?.toUpperCase()}</Tag> {selectedLog.destination_url}
                  </Descriptions.Item>
                  <Descriptions.Item label="状态">
                    <Tag color={selectedLog.status === 'success' ? 'success' : 'error'}>
                      {selectedLog.status === 'success' ? '成功' : '失败'}
                    </Tag>
                  </Descriptions.Item>
                  {selectedLog.error_message && (
                    <Descriptions.Item label="错误信息">
                      <Text type="danger">{selectedLog.error_message}</Text>
                    </Descriptions.Item>
                  )}
                  <Descriptions.Item label="转发内容 (Payload)">
                    <pre style={{ background: '#f5f5f5', padding: '12px', borderRadius: '4px', overflowX: 'auto', maxHeight: '400px', fontSize: '12px' }}>
                      {(() => {
                        try {
                          return JSON.stringify(JSON.parse(selectedLog.payload), null, 2);
                        } catch (e) {
                          return selectedLog.payload;
                        }
                      })()}
                    </pre>
                  </Descriptions.Item>
                </Descriptions>
              )}
            </Modal>
          </Content>
        </Layout>
      </Layout>
    </ConfigProvider >
  );
};

export default App;
