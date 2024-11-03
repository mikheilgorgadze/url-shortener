function copyURL(ele, url) {
    navigator.clipboard.writeText(url)
    console.log(ele)
    ele.textContent = "Copied!"
    window.setTimeout(() => {
        ele.textContent = "Copy"
    }, 1000)
}
