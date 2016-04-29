## Installing JRE

sudo apt-get install -y openjdk-7-jre-headless

## Installing Oracle JDK

sudo apt-get install python-software-properties

sudo add-apt-repository ppa:webupd8team/java

sudo apt-get update

sudo apt-get install oracle-java7-installer

## 修改virtualbox虚拟硬盘位置

使用vagrant创建虚拟机之后，虚拟硬盘默认存储在“C:\Users\\&lt;User&gt;\VirtualBox VMs”。C盘下会有诸多不便，修改路径的方式如下：

* 关闭正在运行的虚拟机，将vmdk文件（如“C:\Users\Administrator\VirtualBox VMs\vagrant_default_1461811043169_4692\box-disk1.vmdk”）复制到其他盘（如D盘）。
* 打开“cmd”，进入到Oracle VM VirtualBox安装目录，执行“VBOXMANAGE.EXE internalcommands sethduuid \<vmdk文件目录\>”，为新的虚拟硬盘文件设置uuid
* 在Oracle VM VirtualBox管理器中选择虚拟机，并进行设置。
* 在设置面板中进入存储设置，可以看到box-disk1.vmdk存储附件，将其删除。
* 为控制器添加新的虚拟硬盘，并选择“使用现有的虚拟盘”，选择新的vmdk文件。
