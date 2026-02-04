package net.favouriteless.elainaocr;

import net.sourceforge.tess4j.TessAPI;

public class Main {

    public static void main(String[] args) {
        System.out.println(TessAPI.INSTANCE.TessVersion());
    }

}
