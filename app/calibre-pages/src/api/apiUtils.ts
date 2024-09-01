// src/api/apiUtils.ts
import {ElNotification} from "element-plus";
import {h} from "vue";

export async function handleApiResponse(response: Response) {
    const data = await response.json();
    if (response.status === 200) {
        if (data.code == 200) {
            return data.data;
        } else {
            ElNotification({
                title: 'ERROR: ' + data.message,
                message: h('i', {style: 'color: red'}, data.message),
                type: 'error'
            })
        }
    } else {
        throw new Error(data.message || 'API request failed');
    }
}