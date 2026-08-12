'use strict';
'require view';
'require form';
'require fs';
'require uci';
'require ui';

return view.extend({
 load: function() {
 return Promise.all([
 uci.load('torrent-parser'),
 fs.exec('/etc/init.d/torrent-parser', [ 'status' ]).catch(function(e) {
 return { code: 1, stdout: '', stderr: String(e || '') };
 })
 ]);
 },

 render: function(data) {
 var running = data[1] && (data[1].code === 0);
 var m = new form.Map('torrent-parser', _('Torrent Parser'),
 _('Лёгкий Torznab-сервис для RuTracker, Kinozal и NNMClub. Настройки хранятся на роутере через UCI.'));

 var s = m.section(form.NamedSection, 'main', 'main', _('Сервис'));
 s.addremove = false;
 s.anonymous = true;

 var o = s.option(form.DummyValue, '_status', _('Состояние'));
 o.cfgvalue = function() { return running? _('Запущен'): _('Остановлен'); };

 o = s.option(form.Flag, 'enabled', _('Включить сервис'));
 o.default = o.enabled;
 o.rmempty = false;

 o = s.option(form.Value, 'listen', _('Адрес прослушивания'));
 o.default = '0.0.0.0:9696';
 o.rmempty = false;
 o.description = _('Например: 0.0.0.0:9696.');

 o = s.option(form.Value, 'api_key', _('API key'));
 o.password = true;
 o.rmempty = false;

 o = s.option(form.Value, 'timeout', _('Тайм-аут HTTP, сек.'));
 o.datatype = 'uinteger';
 o.default = '15';
 o.rmempty = false;

 o = s.option(form.Value, 'proxy', _('Прокси для запросов'));
 o.placeholder = 'socks5://127.0.0.1:1080';
 o.description = _('Необязательно. Поддерживаются http://, https:// и socks5://.');

 o = s.option(form.Value, 'user_agent', _('User-Agent'));
 o.default = 'TorrentParserOpenWrt/0.1';
 o.rmempty = false;

 o = s.option(form.Button, '_restart', _('Управление'));
 o.inputtitle = _('Перезапустить');
 o.inputstyle = 'apply';
 o.onclick = function() {
 return fs.exec('/etc/init.d/torrent-parser', [ 'restart' ]).then(function(res) {
 if (res.code!== 0)
 throw new Error(res.stderr || res.stdout || _('Не удалось перезапустить сервис'));
 ui.addNotification(null, E('p', _('Torrent Parser перезапущен.')));
 window.setTimeout(function() { window.location.reload(); }, 800);
 });
 };

 function trackerSection(name, title, info) {
 var ts = m.section(form.NamedSection, name, 'tracker', title);
 ts.addremove = false;
 ts.anonymous = true;
 if (info)
 ts.description = info;

 var x = ts.option(form.Flag, 'enabled', _('Включить'));
 x.default = x.enabled;
 x.rmempty = false;

 x = ts.option(form.Value, 'base_url', _('Адрес трекера'));
 x.rmempty = false;
 return ts;
 }

 var rt = trackerSection('rutracker', 'RuTracker', _('Поиск и загрузка требуют учётную запись.'));
 o = rt.option(form.Value, 'username', _('Логин'));
 o = rt.option(form.Value, 'password', _('Пароль'));
 o.password = true;

 var kz = trackerSection('kinozal', 'Kinozal', _('Поиск и загрузка требуют учётную запись.'));
 o = kz.option(form.Value, 'username', _('Логин'));
 o = kz.option(form.Value, 'password', _('Пароль'));
 o.password = true;

 var nnm = trackerSection('nnmclub', 'NNMClub', _('Используется анонимный поиск. Поля учётной записи сейчас не используются.'));
 o = nnm.option(form.Value, 'username', _('Логин (зарезервировано)'));
 o = nnm.option(form.Value, 'password', _('Пароль (зарезервировано)'));
 o.password = true;

 var api = m.section(form.TypedSection, '_api', _('Подключение Torznab'));
 api.anonymous = true;
 api.addremove = false;
    o = api.option(form.DummyValue, '_hint', _('URL'));
 o.cfgvalue = function() {
 var host = window.location.hostname;
 var listen = uci.get('torrent-parser', 'main', 'listen') || '0.0.0.0:9696';
 var port = listen.lastIndexOf(':') >= 0
? listen.substring(listen.lastIndexOf(':') + 1)
: '9696';
 var key = uci.get('torrent-parser', 'main', 'api_key') || '';
 return 'http://' + host + ':' + port + '/api?t=search&[REDACTED] + encodeURIComponent(key);
 };

 return m.render();
 }
});
