// plugins/console-banner.client.ts
export default defineNuxtPlugin({
    name: 'console-banner',

    setup() { 
        const { basicConfig } = useSysConfig();
        console.log('\n\x1b[36m-----------------------------------------------------\x1b[0m')
        console.log(`\x1b[32m 🚀 欢迎阅览 ${basicConfig?.value.author} 个人博客\x1b[0m`)
        console.log(`\x1b[33m 📝 站长介绍：${basicConfig?.value.author_desc}\x1b[0m`)
        console.log(`\x1b[34m 🌐 访问地址：${basicConfig?.value.blog_url}\x1b[0m`)
        console.log(`\x1b[32m ✅ 状态：启动成功\x1b[0m`)
        console.log('\x1b[36m-----------------------------------------------------\x1b[0m\n')
    },


});
